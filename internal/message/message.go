package message

import (
	"encoding/json"
	"time"
)

type Message struct {
	Type string				`json:"type"`
	Payload json.RawMessage	`json:"payload"`
	Timestamp int64			`json:"timestamp"`
}

func New(msgType string, payload any) (*Message, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return &Message {
		Type: msgType,
		Payload: data,
		Timestamp: time.Now().UnixMilli(),
	}, nil
}