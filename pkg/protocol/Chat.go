package protocol

type CreateChatRequest struct {
	Name      string `json:"name"`
	PublicKey string `json:"public_key"`
}

type CreateChatResponse struct {
	ChatID string `json:"chat_id"`
}
