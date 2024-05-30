package sdk

import (
	"log"
	"testing"
)

func TestNewSDK(t *testing.T) {
	s, err := NewSDK("http://localhost:8080")
	if err != nil {
		t.Errorf("Error creating SDK: %v", err)
	}

	info, err := s.GetServerInfo()
	if err != nil {
		t.Errorf("Error getting server info: %v", err)
	}

	log.Println(info)
}
