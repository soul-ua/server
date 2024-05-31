package main

import (
	"github.com/soul-ua/server/internal/accounts"
	"github.com/soul-ua/server/internal/webserver"
	"go.etcd.io/bbolt"
)

func main() {
	bdb, err := bbolt.Open(".data/storage.db", 0600, nil)
	if err != nil {
		panic(err)
	}

	accountsUsecase := accounts.NewAccountsBBolt(bdb)
	srv, err := webserver.NewWebserver(accountsUsecase)
	if err != nil {
		panic(err)
	}

	if err := srv.Start(":8080"); err != nil {
		panic(err)
	}
}
