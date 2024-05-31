package accounts

type Accounts interface {
	RegisterAccount(username, publicKey string) error
	GetUserPublicKeyArmor(username string) (string, error)
}
