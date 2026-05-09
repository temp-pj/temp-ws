package room

import (
	"temp-ws/internal/client"
	"temp-ws/internal/message"
)

type Room struct {
	clients map[* client.Client]bool
	register chan *client.Client
	unregister chan *client.Client
	broadcast chan *message.Message
	roomID string
	roomState RoomState
}

func (r *Room) Run() {
	for {
		select {
			case client := <- r.register:
				r.clients[client] = true

			case client := <- r.unregister:
				delete(r.clients, client)
			
			case message := <- r.broadcast:
				for client, _ := range r.clients {
					client.Send <- message
				}
		}
	}
}

func NewRoom() *Room {
	return &Room {
		clients: make(map[* client.Client]bool),
		register: make(chan *client.Client),
		unregister: make(chan *client.Client),
		broadcast: make(chan *message.Message),
	}
}


type RoomState int

const (
	Waiting RoomState = iota
	Playing
	Finished
)

func (s RoomState) String() string {
	switch s {
		case Waiting:
			return "waiting"
		case Playing:
			return "playing"
		case Finished:
			return "finished"
	}

	return "unknown"
}