package protocol

// ContactRequest is client Payload To target SERVER
//
//	YOU->Server: ContactRequest
//	Server->Target: ContactRequested(YOU.username, YOU.PublicKey)
//	Target->YOU: ContactRequestAccepted(Target.PublicKey)
type ContactRequest struct {
	To string `json:"To"`
}

type ContactRequested struct {
	From      string `json:"From"`
	PublicKey string `json:"public_key"`
}

type ContactRequestAccepted struct {
	PublicKey string `json:"public_key"`
}
