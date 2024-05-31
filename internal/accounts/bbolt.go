package accounts

import (
	"errors"
	"go.etcd.io/bbolt"
	"log"
)

var (
	ErrAccountAlreadyExists = errors.New("account already exists")
	ErrorAccountNotFound    = errors.New("account not found")
)

type accountsMemory struct {
	bdb *bbolt.DB
}

var _ Accounts = &accountsMemory{}

func NewAccountsBBolt(bdb *bbolt.DB) Accounts {
	return &accountsMemory{
		bdb: bdb,
	}
}

func (a *accountsMemory) RegisterAccount(username string, publicKey string) error {
	return a.bdb.Update(func(tx *bbolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte("accounts"))
		if err != nil {
			return err
		}

		if bucket.Get([]byte(username)) != nil {
			return ErrAccountAlreadyExists
		}

		return bucket.Put([]byte(username), []byte(publicKey))
	})
}

func (a *accountsMemory) RegisterAccountPrivateKey(username string, privateKey string) error {
	return a.bdb.Update(func(tx *bbolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte("accounts-private"))
		if err != nil {
			return err
		}

		if bucket.Get([]byte(username)) != nil {
			return ErrAccountAlreadyExists
		}

		return bucket.Put([]byte(username), []byte(privateKey))
	})
}

func (a *accountsMemory) GetUserPublicKeyArmor(username string) (string, error) {
	var publicKey []byte
	err := a.bdb.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte("accounts"))
		if bucket == nil {
			return ErrorAccountNotFound
		}

		publicKey = bucket.Get([]byte(username))
		if publicKey == nil {
			log.Println("account not found for username", username)
			return ErrorAccountNotFound
		}

		return nil
	})
	if err != nil {
		return "", err
	}

	return string(publicKey), nil
}

func (a *accountsMemory) GetUserPrivateKeyArmor(username string) (string, error) {
	var privateKey []byte
	err := a.bdb.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte("accounts-private"))
		if bucket == nil {
			return ErrorAccountNotFound
		}
		privateKey = bucket.Get([]byte(username))
		if privateKey == nil {
			log.Println("account not found for username", username)
			return ErrorAccountNotFound
		}

		return nil
	})
	if err != nil {
		return "", err
	}

	return string(privateKey), nil
}
