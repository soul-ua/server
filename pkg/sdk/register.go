package sdk

import (
	"encoding/json"
	"fmt"
	"github.com/soul-ua/server/pkg/protocol"
)

func (s *SDK) Register(username, publicKeyArmor string) error {
	req, _ := json.Marshal(protocol.RegisterRequest{
		Username:  username,
		PublicKey: publicKeyArmor,
	})

	var res protocol.RegisterResponse
	if err := s.Request("POST", "/register", req, &res); err != nil {
		return fmt.Errorf("failed to register: %w", err)
	}

	return nil
}
