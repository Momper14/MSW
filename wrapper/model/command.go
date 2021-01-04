package model

// Command wich for the MSW
type Command struct {
	Target  CommandTarget `json:"target"`
	Payload string        `json:"payload"`
}
