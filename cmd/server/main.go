package main

import (
	"errors"
	"fmt"
	"github.com/soul-ua/server/internal/accounts"
	"github.com/soul-ua/server/internal/webserver"
	"github.com/soul-ua/server/pkg/protocol"
	"go.etcd.io/bbolt"
	"log"
)

func main() {
	bdb, err := bbolt.Open(".data/storage.db", 0600, nil)
	if err != nil {
		panic(err)
	}

	accountsUsecase := accounts.NewAccountsBBolt(bdb)
	_, _, err = ensureServerKeys(accountsUsecase)
	if err != nil {
		panic(err)
	}

	srv, err := webserver.NewWebserver(accountsUsecase)
	if err != nil {
		panic(err)
	}

	if err := srv.Start(":8080"); err != nil {
		panic(err)
	}
}

func ensureServerKeys(accountsUC accounts.Accounts) (string, string, error) {
	var serverPublicKey string
	serverPrivateKey, err := accountsUC.GetUserPrivateKeyArmor("server")
	if errors.Is(err, accounts.ErrorAccountNotFound) {
		log.Println("* generate server keys")
		serverPrivateKey, serverPublicKey, err = protocol.GeneratePair("sever", "server@soul.ua")
		if err != nil {
			return "", "", fmt.Errorf("failed generate server keys: %w", err)
		}

		err = accountsUC.RegisterAccountPrivateKey("server", serverPrivateKey)
		if err != nil {
			return "", "", fmt.Errorf("failed register server private key: %w", err)
		}

		err = accountsUC.RegisterAccount("server", serverPublicKey)
		if err != nil {
			return "", "", fmt.Errorf("failed register server public key: %w", err)
		}

		return serverPrivateKey, serverPublicKey, nil
	} else if err != nil {
		return "", "", fmt.Errorf("failed get server private key: %w", err)
	}

	serverPublicKey, err = accountsUC.GetUserPublicKeyArmor("server")
	if err != nil {
		return "", "", fmt.Errorf("failed get server public key: %w", err)
	}

	return serverPrivateKey, serverPublicKey, nil
}
