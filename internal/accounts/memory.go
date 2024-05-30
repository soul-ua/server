package accounts

import (
	"errors"
	"sync"
)

var (
	ErrAccountAlreadyExists = errors.New("account already exists")
)

type accountsMemory struct {
	mutex    sync.RWMutex
	accounts map[string]string
}

var _ Accounts = &accountsMemory{}

func NewAccountsMemory() Accounts {
	return &accountsMemory{
		accounts: make(map[string]string),
	}
}

func (a *accountsMemory) RegisterAccount(username, publicKeyArmor string) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if _, ok := a.accounts[username]; ok {
		return ErrAccountAlreadyExists
	}

	a.accounts[username] = publicKeyArmor
	return nil
}

func (a *accountsMemory) GetUserPublicKeyArmor(username string) (string, error) {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	key, ok := a.accounts[username]
	if !ok {
		return "", errors.New("user not found")
	}
	return key, nil
}
