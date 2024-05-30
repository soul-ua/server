package protocol

import (
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"log"
)

func Generate(name, email string) {
	ecKey, err := crypto.GenerateKey(name, email, "x25519", 0)
	log.Println(ecKey, err)

	armor, err := ecKey.Armor()
	if err != nil {
		panic(err)
	}

	armorPub, err := ecKey.GetArmoredPublicKey()
	if err != nil {
		panic(err)
	}

	log.Println(armor)

	log.Println(armorPub)
}
