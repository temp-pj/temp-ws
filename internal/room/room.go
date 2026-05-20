package room

import (
	"sync"
	"temp-ws/internal/client"
	"temp-ws/internal/message"
)

type MusicFetcher interface {
	Fetch(category music.Category) ([]music.Song, error)
}


type Room struct {
	clients map[* client.Client]bool
	register chan *client.Client
	unregister chan *client.Client
	broadcast chan *message.Message
	incoming chan *message.ClientMessage
	roomID string
	roomState RoomState
	quit chan struct{}
	closeOnce sync.Once
}

func (r *Room) Run() {
	for {
		select {
			case client := <- r.register:
				if client == nil || client.Send == nil { continue }
				r.clients[client] = true

				msg, err := message.New(message.TypePlayerJoined, message.PlayerJoinedPayload { PlayerID: client.ID })
				if err != nil { continue }
				
				for c := range r.clients {
					select {
						case c.Send <- msg:

						default:
							delete(r.clients, c)
							close(c.Send)
					}
					
				}

			case client := <- r.unregister:
				delete(r.clients, client)
				close(client.Send)

				msg, err := message.New(message.TypePlayerLeft, message.PlayerLeftPayload { PlayerID: client.ID })
				if err != nil { continue }

				for c := range r.clients {
					select {
						case c.Send <- msg:

						default:
							delete(r.clients, c)
							close(c.Send)
					}
				}
			
			case message := <- r.broadcast:
				for client := range r.clients {
					select {
						case client.Send <- message:

						default: 
							delete(r.clients, client)
							close(client.Send)

					}
				}

			case clientMessage := <- r.incoming:
					actions, cleanup := r.HandleClientMessage(clientMessage)
					r.executeActions(actions)

					if cleanup != nil { cleanup() }
			case <- r.quit: return 
			

		}
	}
}

func (r *Room) Close() {
	  r.closeOnce.Do(func() { close(r.quit) })
}

func (r *Room) Register(c *client.Client) {
	r.register <- c
}

func (r *Room) Unregister(c *client.Client) {
	r.unregister <- c
}

func (r *Room) Broadcast(msg *message.Message) {
	r.broadcast <- msg
}

func NewRoom(roomID string) *Room {
func NewRoom(roomID string, musicFetcher MusicFetcher) *Room {
	return &Room {
		clients: make(map[string]*client.Client),
		register: make(chan *client.Client),
		unregister: make(chan *client.Client),
		broadcast: make(chan *message.Message),
		incoming: make(chan *message.ClientMessage),
		roomID: roomID,
		roomState: Waiting,
		quit: make(chan struct{}),
	}
}


type RoomState int

const (
	Waiting RoomState = iota
	Playing
	Finished
	Closed
)

func (s RoomState) String() string {
	switch s {
		case Waiting:
			return "waiting"
		case Playing:
			return "playing"
		case Finished:
			return "finished"
		case Closed:
			return "closed"
	}

	return "unknown"
}