package model

// Message wich the MSW sends
type Message struct {
	Type    MessageType `json:"type"`
	Payload string      `json:"payload"`
}
