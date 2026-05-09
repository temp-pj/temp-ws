package message

import "encoding/json"

type Message struct {
	Type string
	Payload json.RawMessage
	Timestamp int64
}