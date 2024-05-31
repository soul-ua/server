package sdk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/soul-ua/server/pkg/protocol"
	"io"
	"log"
	"net/http"
)

type Keychain interface {
	SavePublicKey(username, privateKeyArmor string) error
	GetPublicKey(username string) (string, error)
}

type SDK struct {
	serverURL string
	info      protocol.ServerInfo
	keychain  Keychain

	username   string
	privateKey *crypto.Key
}

func NewSDKArmor(serverURL string, keychain Keychain, username, privateKeyArmor string) (*SDK, error) {
	key, err := crypto.NewKeyFromArmored(privateKeyArmor)
	if err != nil {
		return nil, fmt.Errorf("failed to load private key: %w", err)
	}

	return NewSDK(serverURL, keychain, username, key)
}

func NewSDK(serverURL string, keychain Keychain, username string, key *crypto.Key) (*SDK, error) {
	s := &SDK{
		serverURL: serverURL,
		keychain:  keychain,

		username:   username,
		privateKey: key,
	}

	info, err := s.GetServerInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to get server info: %w", err)
	}

	s.info = info

	return s, nil
}

func (s *SDK) GetServerInfo() (protocol.ServerInfo, error) {
	rsp, err := http.Get(s.serverURL + "/.soul-server-info")
	if err != nil {
		return protocol.ServerInfo{}, fmt.Errorf("failed to get server info: %w", err)
	}

	body, err := io.ReadAll(rsp.Body)
	if err != nil {
		return protocol.ServerInfo{}, fmt.Errorf("failed to read server info: %w", err)
	}

	pgpSignatureBase64 := rsp.Header.Get("PGP-Signature")

	var info protocol.ServerInfo
	if err = json.Unmarshal(body, &info); err != nil {
		return protocol.ServerInfo{}, fmt.Errorf("failed to decode server info: %w", err)
	}

	if err = protocol.VerifySignArmor(body, pgpSignatureBase64, info.PublicKey); err != nil {
		return protocol.ServerInfo{}, fmt.Errorf("failed to verify server info: %w", err)
	}

	return info, nil
}

func (s *SDK) Request(method, path string, data []byte) ([]byte, error) {
	r, err := http.NewRequest(method, s.serverURL+path, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	pgpSignatureBase64, err := protocol.Sign(data, s.privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign data: %w", err)
	}

	r.Header.Add("soul-username", s.username)
	r.Header.Add("PGP-Signature", pgpSignatureBase64)

	log.Println("req sent")
	rsp, err := http.DefaultClient.Do(r)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer rsp.Body.Close()
	log.Println("req done")

	body, err := io.ReadAll(rsp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}
	log.Println("req read done")

	pgpSignatureBase64 = rsp.Header.Get("PGP-Signature")

	if err = protocol.VerifySignArmor(body, pgpSignatureBase64, s.info.PublicKey); err != nil {
		return nil, fmt.Errorf("failed to verify response sign: %w", err)
	}

	log.Println("req verify done")

	return body, nil
}

// SendEnvelope just send envelope to the server
func (s *SDK) SendEnvelope(envelop *protocol.Envelope) error {
	envelop.From = s.username
	packed, err := envelop.Pack()
	if err != nil {
		return fmt.Errorf("failed to pack envelope: %w", err)
	}

	rsp, err := s.Request("POST", "/send", packed)
	if err != nil {
		return fmt.Errorf("failed to send envelope: %w", err)
	}

	log.Println("Send envelope response:", string(rsp))
	return nil
}
