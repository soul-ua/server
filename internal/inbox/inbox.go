package inbox

import (
	"fmt"
	"github.com/google/uuid"
	"go.etcd.io/bbolt"

	"github.com/soul-ua/server/pkg/protocol"
)

type Inbox struct {
	bdb      *bbolt.DB
	username string
}

func NewInbox(username string) (*Inbox, error) {
	bdb, err := bbolt.Open(fmt.Sprintf(".data/inbox-%s.db", username), 0600, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to open inbox database: %w", err)
	}

	inbx := &Inbox{
		bdb:      bdb,
		username: username,
	}

	return inbx, nil
}

func (i *Inbox) Close() error {
	return i.bdb.Close()
}

func (i *Inbox) Append(envelope *protocol.Envelope) (uuid.UUID, error) {
	envelopeID, err := uuid.NewV7() // should be v7 for binary sort
	if err != nil {
		panic(err)
	}

	if envelope.To != i.username {
		return envelopeID, fmt.Errorf("envelope is not addressed to this inbox")
	}

	packed, err := envelope.Pack()
	if err != nil {
		return envelopeID, fmt.Errorf("failed to pack envelope: %w", err)
	}

	err = i.bdb.Update(func(tx *bbolt.Tx) error {
		mailbox, err := tx.CreateBucketIfNotExists([]byte("mailbox"))
		if err != nil {
			return fmt.Errorf("failed to create mailbox bucket: %w", err)
		}

		return mailbox.Put([]byte(envelopeID.String()), packed)
	})

	if err != nil {
		return envelopeID, fmt.Errorf("failed to write envelope to mailbox: %w", err)
	}

	return envelopeID, nil
}

func (i *Inbox) Read(since []byte, cb func(payload []byte) error) error {
	return i.bdb.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte("mailbox"))
		if bucket == nil {
			return nil
		}
		c := bucket.Cursor()
		if c == nil {
			return nil
		}

		k, v := c.Seek(since)
		limit := 100
		for {
			if k == nil {
				break
			}

			if err := cb(v); err != nil {
				return err
			}

			limit--
			if limit <= 0 {
				break
			}
			k, v = c.Next()
		}
		return nil
	})
}
