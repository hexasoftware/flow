package flowmsg

import "encoding/json"

// RecvMessage Main message structure
type RecvMessage struct {
	OP   string          `json:"op"`
	ID   string          `json:"id"` // destination sessId
	Data json.RawMessage `json:"data"`
}

// SendMessage
type SendMessage struct {
	OP   string      `json:"op"`
	ID   string      `json:"id"`
	Data interface{} `json:"data"`
}
