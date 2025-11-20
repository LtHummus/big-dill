package socket

type MessageWrapper struct {
	Source  *Client
	Message *Message
}

type Message struct {
	Kind    string         `json:"kind"`
	Payload map[string]any `json:"payload"`
}
