package webserver

import (
	"encoding/json"
	"fmt"
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/soul-ua/server/internal/accounts"
	"github.com/soul-ua/server/pkg/protocol"
	"io"
	"net/http"
	"os"
)

type Webserver struct {
	accounts accounts.Accounts

	publicKey          string
	privateKey         string
	unlockedPrivateKey *crypto.Key
}

func NewWebserver(accountsUC accounts.Accounts) (*Webserver, error) {
	privateKey, err := os.ReadFile("./.keys/server.prv")
	if err != nil {
		return nil, fmt.Errorf("failed to read private key: %w", err)
	}

	publicKey, err := os.ReadFile("./.keys/server.pub")
	if err != nil {
		return nil, fmt.Errorf("failed to read public key: %w", err)
	}

	privateKeyObj, err := crypto.NewKeyFromArmored(string(privateKey))
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	return &Webserver{
		accounts: accountsUC,

		publicKey:          string(publicKey),
		privateKey:         string(privateKey),
		unlockedPrivateKey: privateKeyObj,
	}, nil
}

func (w *Webserver) Start(addr string) error {
	fmt.Println("Starting webserver on", addr)

	http.HandleFunc("GET /.soul-server-info", w.handleServerInfo)

	http.HandleFunc("POST /register", w.handleRegister)

	return http.ListenAndServe(addr, nil)
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
}

func (w *Webserver) handleServerInfo(wr http.ResponseWriter, r *http.Request) {
	data, _ := json.Marshal(protocol.ServerInfo{
		Version:         "0.0.0",
		PublicKey:       w.publicKey,
		CurrentUnitTime: crypto.GetUnixTime(),
	})
	_ = w.sendSign(data, wr)
}

func (w *Webserver) decodeVerifyUserRequest(r *http.Request, v interface{}) error {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("failed to read body: %w", err)
	}

	username := r.Header.Get("soul-username")
	pgpSignatureBase64 := r.Header.Get("PGP-Signature")

	userPublicKeyArmor, err := w.accounts.GetUserPublicKeyArmor(username)
	if err != nil {
		return fmt.Errorf("failed to get user public key: %w", err)
	}

	if err := protocol.VerifySignArmor(data, pgpSignatureBase64, userPublicKeyArmor); err != nil {
		return fmt.Errorf("failed to verify signature: %w", err)
	}

	return json.Unmarshal(data, v)
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
