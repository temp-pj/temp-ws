package message

import (
	"encoding/json"
	"time"
)

type Message struct {
	Type string
	Payload json.RawMessage
	Timestamp int64
}

func New(msgType string, payload any) *Message {
	data, _ := json.Marshal(payload)

	return &Message {
		Type: msgType,
		Payload: data,
		Timestamp: time.Now().UnixMilli(),
	}
}