package room

import "temp-ws/internal/message"

type Action struct {
	Type ActionType
	Target string
	Message *message.Message
}

type ActionType int 


const (
	Broadcast ActionType = iota
	Unicast
)