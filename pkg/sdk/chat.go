package sdk

import (
	"encoding/json"
	"github.com/soul-ua/server/pkg/protocol"
)

func (s *SDK) CreateChat(name string) (string, string, error) {
	chatPrivate, chatPublic, err := protocol.GeneratePair(name, "")
	if err != nil {
		return "", "", err
	}
	req, _ := json.Marshal(protocol.CreateChatRequest{
		Name:      name,
		PublicKey: chatPublic,
	})

	var res protocol.CreateChatResponse
	body, err := s.Request("PUT", "/chat", req)
	if err != nil {
		return "", "", err
	}

	if err = json.Unmarshal(body, &res); err != nil {
		return "", "", err
	}

	return res.ChatID, chatPrivate, nil
}
