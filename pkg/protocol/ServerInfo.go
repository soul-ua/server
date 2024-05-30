package protocol

type ServerInfo struct {
	Version         string
	URL             string
	PublicKey       string
	CurrentUnitTime int64
}
