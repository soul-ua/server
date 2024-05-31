package chat

import (
	"fmt"
	"github.com/google/uuid"
	"go.etcd.io/bbolt"
)

type Chat struct {
	bdb    *bbolt.DB
	chatID string
}

func NewChat(chatID string) (*Chat, error) {
	if err := uuid.Validate(chatID); err != nil {
		return nil, fmt.Errorf("invalid chat ID: %w", err)
	}

	bdb, err := bbolt.Open(fmt.Sprintf(".data/chat-%s.db", chatID), 0600, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to open chat database: %w", err)
	}

	chat := &Chat{
		bdb:    bdb,
		chatID: chatID,
	}

	return chat, nil
}

func CreateChat(creatorUsername, name, publicKey string) (*Chat, error) {
	chatID := uuid.New().String()
	chat, err := NewChat(chatID)
	if err != nil {
		return nil, err
	}
	err = chat.bdb.Update(func(tx *bbolt.Tx) error {
		metadata, err := tx.CreateBucketIfNotExists([]byte("metadata"))
		if err != nil {
			return err
		}

		if metadata.Put([]byte("name"), []byte(name)) != nil {
			return fmt.Errorf("failed to put name into metadata")
		}

		if metadata.Put([]byte("creator"), []byte(creatorUsername)) != nil {
			return fmt.Errorf("failed to put creator into metadata")
		}

		if metadata.Put([]byte("publicKey"), []byte(publicKey)) != nil {
			return fmt.Errorf("failed to put publicKey into metadata")
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	return chat, nil
}

func (c *Chat) Close() error {
	return c.bdb.Close()
}

func (c *Chat) GetChatID() string {
	return c.chatID
}
