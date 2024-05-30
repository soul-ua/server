package accounts

type Accounts interface {
	RegisterAccount(username, publicKeyArmor string) error
	GetUserPublicKeyArmor(username string) (string, error)
}
