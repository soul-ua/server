package protocol

type GetInboxRequest struct {
	SinceID string `json:"since_id"`
}

// GetInboxResponse is server packed envelopes with is gob encoded
type GetInboxResponse struct {
	Envelopes [][]byte
}
