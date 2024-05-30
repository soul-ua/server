package protocol

type RegisterRequest struct {
	Username  string `json:"username"`
	PublicKey string `json:"public_key"`
}

type RegisterResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}
