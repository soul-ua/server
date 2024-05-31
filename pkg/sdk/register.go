package sdk

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"github.com/soul-ua/server/pkg/protocol"
	"log"
)

func (s *SDK) Register(username, publicKeyArmor string) error {
	req, _ := json.Marshal(protocol.RegisterRequest{
		Username:  username,
		PublicKey: publicKeyArmor,
	})

	var res protocol.RegisterResponse
	body, err := s.Request("POST", "/register", req)
	if err != nil {
		return fmt.Errorf("failed to register: %w", err)
	}

	if err = json.Unmarshal(body, &res); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	return nil
}

func (s *SDK) ContactRequest(username string) error {
	req, _ := json.Marshal(protocol.ContactRequest{
		To: username,
	})

	_, err := s.Request("POST", "/contact/request", req)
	if err != nil {
		return fmt.Errorf("failed to send contact request: %w", err)
	}

	return nil
}

func (s *SDK) GetInbox(sinceID string) ([]*protocol.Envelope, error) {
	req, _ := json.Marshal(protocol.GetInboxRequest{
		SinceID: sinceID,
	})
	body, err := s.Request("POST", "/inbox", req)
	if err != nil {
		return nil, fmt.Errorf("failed to get inbox: %w", err)
	}

	log.Println("inbox received")

	var res protocol.GetInboxResponse
	responseBuf := bytes.NewBuffer(body)
	dec := gob.NewDecoder(responseBuf)
	if err = dec.Decode(&res); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	result := make([]*protocol.Envelope, len(res.Envelopes))
	for i, envelopePacked := range res.Envelopes {
		envelope, err := protocol.UnpackEnvelope(envelopePacked)
		if err != nil {
			return nil, fmt.Errorf("failed to unpack envelope[%d]: %w", i, err)
		}
		result[i] = envelope
	}

	return result, nil
}
