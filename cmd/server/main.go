package main

import (
	"github.com/soul-ua/server/internal/accounts"
	"github.com/soul-ua/server/internal/webserver"
)

func main() {
	accountsUsecase := accounts.NewAccountsMemory()
	srv, err := webserver.NewWebserver(accountsUsecase)
	if err != nil {
		panic(err)
	}

	if err := srv.Start(":8080"); err != nil {
		panic(err)
	}
}
