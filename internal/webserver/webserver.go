package webserver

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/soul-ua/server/internal/accounts"
	"github.com/soul-ua/server/internal/chat"
	"github.com/soul-ua/server/internal/inbox"
	"github.com/soul-ua/server/pkg/protocol"
	"io"
	"log"
	"net/http"
)

type Webserver struct {
	accounts accounts.Accounts

	publicKey          string
	privateKey         string
	unlockedPrivateKey *crypto.Key
}

func NewWebserver(accountsUC accounts.Accounts) (*Webserver, error) {
	privateKey, err := accountsUC.GetUserPrivateKeyArmor("server")
	if err != nil {
		return nil, fmt.Errorf("failed to read private key: %w", err)
	}

	publicKey, err := accountsUC.GetUserPublicKeyArmor("server")
	if err != nil {
		return nil, fmt.Errorf("failed to read public key: %w", err)
	}

	privateKeyObj, err := crypto.NewKeyFromArmored(privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	return &Webserver{
		accounts: accountsUC,

		publicKey:          publicKey,
		privateKey:         privateKey,
		unlockedPrivateKey: privateKeyObj,
	}, nil
}

func (w *Webserver) Start(addr string) error {
	fmt.Println("Starting webserver on", addr)

	http.HandleFunc("GET /.soul-server-info", w.handleServerInfo)

	http.HandleFunc("POST /register", w.handleRegister)
	http.HandleFunc("POST /contact/request", w.handleContactRequest)
	http.HandleFunc("POST /inbox", w.handleInboxRequest)
	http.HandleFunc("POST /send", w.handleSend)

	http.HandleFunc("PUT /chat", w.handleCreateChat)

	return http.ListenAndServe(addr, nil)
}

func (w *Webserver) handleInboxRequest(wr http.ResponseWriter, r *http.Request) {
	var req protocol.GetInboxRequest
	username, err := w.decodeVerifyUserRequest(r, &req)
	if err != nil {
		panic(err)
	}

	log.Printf("[%s] get inbox since: %s", username, req.SinceID)

	inbx, err := inbox.NewInbox(username)
	if err != nil {
		panic(err)
	}
	defer inbx.Close()

	envelopes := make([][]byte, 0)

	var sinceID []byte
	sinceID = nil // todo: implement me
	err = inbx.Read(sinceID, func(payload []byte) error {
		envelopes = append(envelopes, payload)
		return nil
	})
	if err != nil {
		panic(err)
	}

	res := bytes.Buffer{}
	enc := gob.NewEncoder(&res)
	err = enc.Encode(protocol.GetInboxResponse{
		Envelopes: envelopes,
	})
	if err != nil {
		panic(err)
	}

	_ = w.sendSign(res.Bytes(), wr)
}

func (w *Webserver) handleContactRequest(wr http.ResponseWriter, r *http.Request) {
	// WARNING: This code is vulnerable men in the middle attack (men - server)
	//     this is not zero-trust architecture
	//
	//     userA who send this request, should add their sign of his public key, so userB can verify it and trust

	var req protocol.ContactRequest
	username, err := w.decodeVerifyUserRequest(r, &req)
	if err != nil {
		panic(err)
	}

	userPublicKey, err := w.accounts.GetUserPublicKeyArmor(req.To)
	if err != nil {
		panic(err)
	}

	payload, err := protocol.EncryptStructSign(protocol.ContactRequested{
		From:      username,
		PublicKey: userPublicKey,
	}, userPublicKey, w.privateKey)
	if err != nil {
		panic(err)
	}

	envelope := &protocol.Envelope{
		From:        "server",
		To:          req.To,
		PayloadType: "ContactRequested",
		Payload:     payload,
	}

	inbx, err := inbox.NewInbox(req.To)
	if err != nil {
		panic(err)
	}
	defer inbx.Close()

	if _, err := inbx.Append(envelope); err != nil {
		panic(err)
	}

	_ = w.sendSign([]byte(`{"success":true}`), wr)
}

func (w *Webserver) handleCreateChat(wr http.ResponseWriter, r *http.Request) {
	var req protocol.CreateChatRequest
	username, err := w.decodeVerifyUserRequest(r, &req)
	if err != nil {
		panic(err)
	}

	log.Println(username, "create chat", req)
	c, err := chat.CreateChat(username, req.Name, req.PublicKey)
	if err != nil {
		panic(err)
	}
	defer c.Close()

	res, _ := json.Marshal(protocol.CreateChatResponse{
		ChatID: c.GetChatID(),
	})
	_ = w.sendSign(res, wr)
}

func (w *Webserver) handleSend(wr http.ResponseWriter, r *http.Request) {
	username, body, err := w.verifyUserRequest(r)
	if err != nil {
		panic(err)
	}

	envelope, err := protocol.UnpackEnvelope(body)
	if err != nil {
		panic(err)
	}

	if envelope.From != username {
		panic("username in body and header are not equal")
	}

	// todo: check is current user in contact list of to
	// todo: what if payload encrypted with wrong key? O_o how to check it?

	inbx, err := inbox.NewInbox(envelope.To)
	if err != nil {
		panic(err)
	}
	defer inbx.Close()

	if _, err := inbx.Append(envelope); err != nil {
		panic(err)
	}

	wr.WriteHeader(http.StatusOK)
}

func (w *Webserver) handleRegister(wr http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}
	var registerRequest protocol.RegisterRequest
	if err := json.Unmarshal(data, &registerRequest); err != nil {
		panic(err)
	}

	username := r.Header.Get("soul-username")

	if registerRequest.Username != username {
		panic("username in body and header are not equal")
	}

	pgpSignatureBase64 := r.Header.Get("PGP-Signature")

	if err := protocol.VerifySignArmor(data, pgpSignatureBase64, registerRequest.PublicKey); err != nil {
		panic(fmt.Errorf("failed to verify signature: %w", err))
	}

	if err := w.accounts.RegisterAccount(registerRequest.Username, registerRequest.PublicKey); err != nil {
		panic(err)
	}

	res, _ := json.Marshal(protocol.RegisterResponse{
		Success: true,
	})
	_ = w.sendSign(res, wr)
}

func (w *Webserver) handleServerInfo(wr http.ResponseWriter, r *http.Request) {
	data, _ := json.Marshal(protocol.ServerInfo{
		Version:         "0.0.0",
		PublicKey:       w.publicKey,
		CurrentUnitTime: crypto.GetUnixTime(),
	})
	_ = w.sendSign(data, wr)
}

func (w *Webserver) decodeVerifyUserRequest(r *http.Request, v interface{}) (string, error) {
	username, data, err := w.verifyUserRequest(r)
	if err != nil {
		return "", fmt.Errorf("failed to verify user request: %w", err)
	}

	return username, json.Unmarshal(data, v)
}

func (w *Webserver) verifyUserRequest(r *http.Request) (string, []byte, error) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		return "", nil, fmt.Errorf("failed to read body: %w", err)
	}

	username := r.Header.Get("soul-username")
	pgpSignatureBase64 := r.Header.Get("PGP-Signature")

	userPublicKeyArmor, err := w.accounts.GetUserPublicKeyArmor(username)
	if err != nil {
		return "", nil, fmt.Errorf("failed to get user public key: %w", err)
	}

	if err := protocol.VerifySignArmor(data, pgpSignatureBase64, userPublicKeyArmor); err != nil {
		return "", nil, fmt.Errorf("failed to verify signature: %w", err)
	}

	return username, data, nil
}

func (w *Webserver) sendSign(data []byte, wr http.ResponseWriter) error {
	pgpSignatureBase64, err := protocol.Sign(data, w.unlockedPrivateKey)
	if err != nil {
		return fmt.Errorf("failed to sign data: %w", err)
	}

	wr.Header().Set("Content-Type", "application/json")
	wr.Header().Set("PGP-Signature", pgpSignatureBase64)

	_, _ = wr.Write(data)

	return nil
}
