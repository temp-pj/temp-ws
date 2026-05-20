package room

import (
	"encoding/json"
	"sync"
	"temp-ws/internal/client"
	"temp-ws/internal/game"
	"temp-ws/internal/message"
	"temp-ws/internal/music"
)

type MusicFetcher interface {
	Fetch(category music.Category) ([]music.Song, error)
}


type Room struct {
	clients map[string]*client.Client
	register chan *client.Client
	unregister chan *client.Client
	broadcast chan *message.Message
	incoming chan *message.ClientMessage
	internal chan func() ([]Action, func())
	roomID string
	hostID string
	roomState RoomState
	ready map[string]bool
	quit chan struct{}
	closeOnce sync.Once
	game *game.Game
	musicFetcher MusicFetcher
}

func (r *Room) Run() {
	for {
		select {
			case client := <- r.register:
				if client == nil || client.Send == nil { continue }
				r.clients[client.ID] = client

				msg, err := message.New(message.TypePlayerJoined, message.PlayerJoinedPayload { PlayerID: client.ID })
				if err != nil { continue }
				
				for _, c := range r.clients {
					select {
						case c.Send <- msg:

						default:
							delete(r.clients, c.ID)
							close(c.Send)
					}
				}

			case client := <- r.unregister:
				delete(r.clients, client.ID)
				close(client.Send)

				msg, err := message.New(message.TypePlayerLeft, message.PlayerLeftPayload { PlayerID: client.ID })
				if err != nil { continue }

				for _, c := range r.clients {
					select {
						case c.Send <- msg:

						default:
							delete(r.clients, c.ID)
							close(c.Send)
					}
				}
			
			case msg := <- r.broadcast:
				for _, c := range r.clients {
					select {
						case c.Send <- msg:

						default: 
							delete(r.clients, c.ID)
							close(c.Send)

					}
				}

			case clientMessage := <- r.incoming:
					actions, cleanup := r.HandleClientMessage(clientMessage)
					r.executeActions(actions)

					if cleanup != nil { cleanup() }

			case function := <- r.internal:
				actions, cleanup := function()
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

func (r *Room) IncomingChannel() chan *message.ClientMessage {
	return r.incoming
}

func (r *Room) HandleClientMessage(msg *message.ClientMessage) ([]Action, func()) {
	switch msg.Message.Type {
		case "START_GAME":
			var payload message.StartGamePayload

			err := json.Unmarshal(msg.Message.Payload, &payload)
			if err != nil { return nil, nil }

			playerIDs := make([]string, 0, len(r.clients))
			for id := range r.clients {
				playerIDs = append(playerIDs, id)
			}
			
			category := payload.Category
			
			go func() {
				songs, _ := r.musicFetcher.Fetch(category)
				
				r.internal <- func() ([]Action, func()) {
					r.game = game.NewGame(playerIDs, songs)
					url := r.game.CurrentSong().WrappedURL()

					preloadSongPayload := message.PreloadSongPayload { URL: url }
					preloadSongMsg, _ := message.New("PRELOAD_SONG", preloadSongPayload)

					return []Action {
						{ Type: Broadcast, Message: preloadSongMsg },
					}, nil
				}
			}()
			
			r.roomState = Playing
			gameStartedMsg, _ := message.New("GAME_STARTED", nil)

			return []Action{ {Type: Broadcast, Message: gameStartedMsg }, }, nil

		case "READY_TO_PLAY":
			if r.game == nil { return nil, nil }

			r.ready[msg.From] = true
			allReady := true

			for id := range r.clients {
				if !r.ready[id] {  
					allReady = false
					break
				}
			}

			if !allReady { return nil, nil }

			r.ready = make(map[string]bool)

			r.game.StartRound(func() {
				r.internal <- func() ([]Action, func()) {
					return r.endRoundActions("")
				}
			})
			
			roundStartPayload := message.RoundStartPayload { LetterCards: r.game.CurrentRoundData().LetterCards }
			roundStartMsg, _ := message.New("ROUND_START", roundStartPayload)

			return []Action { { Type: Broadcast, Message: roundStartMsg } }, nil


		case "KICK_PLAYER":
			if msg.From != r.hostID { return nil, nil }
			
			var payload message.KickPlayerPayload

			err := json.Unmarshal(msg.Message.Payload, &payload)
			if err != nil { return nil, nil }

			target := payload.TargetPlayerID

			if _, ok := r.clients[target]; !ok { return nil, nil }

			kickPlayerMsg, _ := message.New("KICKED", nil)
		
			return []Action { { Type: Unicast, Target: target, Message: kickPlayerMsg } }, func() { 
				if c, ok := r.clients[target]; ok {
					close(c.Send)
					delete(r.clients, c.ID)
				}

				delete(r.ready, target)
			 }
		
		case "SUBMIT_ANSWER":
			if r.game == nil { return nil, nil }
			if r.game.CurrentRoundData().State != game.Playing { return nil, nil }

			var payload message.SubmitAnswerPayload
			err := json.Unmarshal(msg.Message.Payload, &payload)
			if err != nil { return nil, nil }
			
			answer := payload.Answer
			correct := r.game.SubmitAnswer(answer)

			if !correct {
				wrongAnswerPayload := message.WrongAnswerPayload { WrongAnswer: answer }
				wrongAnswerMsg, _ := message.New("WRONG_ANSWER", wrongAnswerPayload)

				return []Action{
					{Type: Unicast, Target: msg.From, Message: wrongAnswerMsg },
				}, nil
			}

			return r.endRoundActions(msg.From)
	}

	return nil, nil
}

func (r *Room) endRoundActions(winner string) ([]Action, func()) {
    r.game.EndRound(winner)

    roundResultMsg, _ := message.New("ROUND_RESULT", message.RoundResultPayload{
        Winner: winner, Scores: r.game.Scores(),
    })

    if r.game.NextRound() {
        preloadMsg, _ := message.New("PRELOAD_SONG", message.PreloadSongPayload{
            URL: r.game.CurrentSong().WrappedURL(),
        })
        return []Action{
            {Type: Broadcast, Message: roundResultMsg},
            {Type: Broadcast, Message: preloadMsg},
        }, nil
    }

    gameOverMsg, _ := message.New("GAME_OVER", message.GameOverPayload{
        Score: r.game.Scores(),
    })

	r.roomState = Finished

    return []Action{
		{Type: Broadcast, Message: roundResultMsg},
        {Type: Broadcast, Message: gameOverMsg},
    }, func() { r.game = nil }
}

func (r *Room) executeActions(actions []Action) {
	for _, a := range actions {
		switch a.Type {
			case Broadcast:
				for id, c := range r.clients {
					select {
					case c.Send <- a.Message:
					default:
						delete(r.clients, id)
						close(c.Send)
					}
					
				}

			case Unicast:
				if target, ok := r.clients[a.Target]; ok {
					select {
						case target.Send <- a.Message:
						default:
							delete(r.clients, a.Target)
							close(target.Send)
					}
					
				}

		}
	}
}

func NewRoom(roomID string, musicFetcher MusicFetcher) *Room {
	return &Room {
		clients: make(map[string]*client.Client),
		register: make(chan *client.Client),
		unregister: make(chan *client.Client),
		broadcast: make(chan *message.Message),
		incoming: make(chan *message.ClientMessage),
		internal: make(chan func()([]Action, func())),
		roomID: roomID,
		roomState: Waiting,
		ready: make(map[string]bool),
		quit: make(chan struct{}),
		musicFetcher: musicFetcher,
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