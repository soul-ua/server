package accounts

type Accounts interface {
	RegisterAccount(username, publicKey string) error
	RegisterAccountPrivateKey(username, privateKey string) error

	GetUserPublicKeyArmor(username string) (string, error)
	GetUserPrivateKeyArmor(username string) (string, error)
}
