package room

import (
	"context"
	"encoding/json"
	"sync"
	"temp-ws/internal/client"
	"temp-ws/internal/game"
	"temp-ws/internal/message"
	"temp-ws/internal/music"
	"time"
)

type MusicProvider interface {
	FetchSongs(ctx context.Context, category music.Category, limit int) ([]music.Song, error)
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
	state RoomState
	ready map[string]bool
	quit chan struct{}
	closeOnce sync.Once
	game *game.Game
	musicProvider MusicProvider
}

func (r *Room) Run() {
	ticker := time.NewTicker(1 * time.Second)

	defer ticker.Stop()

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


			case <- ticker.C:
				if r.game == nil { continue }
				if !r.IsPlaying() { continue }

				remaining := r.game.RemainingTime()
				msg, err := message.New("COUNTDOWN", message.CountDownPayload { Remaining: remaining })
				if err != nil { continue }

				for _, c := range r.clients {
					select {
					case c.Send <- msg:
					default:
						delete(r.clients, c.ID)
						close(c.Send)
					}
				}

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
			trackCount := payload.TrackCount

			if trackCount <= 0 { trackCount = 50 }
			
			go func() {
				songs, err := r.musicProvider.FetchSongs(context.Background(), category, trackCount)
				
				if err != nil || len(songs) == 0 { return }
				
				r.internal <- func() ([]Action, func()) {
					r.game = game.NewGame(playerIDs, songs, payload.TimeLimit)
					isrc := r.game.CurrentISRC()
					preloadSongPayload := message.PreloadSongPayload { ISRC: isrc }
					preloadSongMsg, _ := message.New("PRELOAD_SONG", preloadSongPayload)

					return []Action {
						{ Type: Broadcast, Message: preloadSongMsg },
					}, nil
				}
			}()
			
			r.state = Playing
			gameStartedMsg, _ := message.New("GAME_STARTED", nil)

			return []Action{ {Type: Broadcast, Message: gameStartedMsg }, }, nil

		case "READY_TO_PLAY":
			if r.game == nil { return nil, nil }

			var payload message.ReadyToPlayPayload
			err := json.Unmarshal(msg.Message.Payload, &payload)

			if err != nil { return nil, nil }

			if payload.RoundNumber != r.game.CurrentRound() { return nil, nil }

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

			currentRound := r.game.CurrentRound()

			r.game.StartRound(func() {
				r.internal <- func() ([]Action, func()) {
					if r.game.CurrentRound() != currentRound {
						return nil, nil
					}

					return r.endRoundActions("")
				}
			})
			
			roundStartInfo := r.game.GetRoundStartInfo()

			roundStartPayload := message.RoundStartPayload { 
				RoundNumber: roundStartInfo.RoundNumber, 
				TotalRound: roundStartInfo.TotalRounds, 
				LetterCards: roundStartInfo.LetterCards,
			}

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
			if !r.game.IsPlaying() { return nil, nil }

			var payload message.SubmitAnswerPayload
			err := json.Unmarshal(msg.Message.Payload, &payload)
			if err != nil { return nil, nil }
			
			answer := payload.Answer
			correct := r.game.SubmitAnswer(answer)

			if !correct {
				wrongAnswerPayload := message.WrongAnswerPayload { PlayerID: msg.From, WrongAnswer: answer }
				wrongAnswerMsg, _ := message.New("WRONG_ANSWER", wrongAnswerPayload)

				return []Action{
					{Type: Broadcast, Message: wrongAnswerMsg },
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
		isrc := r.game.CurrentISRC()
        preloadMsg, _ := message.New("PRELOAD_SONG", message.PreloadSongPayload{
            ISRC: isrc,
        })

        return []Action{
            {Type: Broadcast, Message: roundResultMsg},
            {Type: Broadcast, Message: preloadMsg},
        }, nil
    }

    gameOverMsg, _ := message.New("GAME_OVER", message.GameOverPayload{
        Score: r.game.Scores(),
    })

	r.state = Finished

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

func (r *Room) IsPlaying()bool {
	return r.state == Playing
}

func NewRoom(roomID string, musicProvider MusicProvider) *Room {
	return &Room {
		clients: make(map[string]*client.Client),
		register: make(chan *client.Client),
		unregister: make(chan *client.Client),
		broadcast: make(chan *message.Message),
		incoming: make(chan *message.ClientMessage),
		internal: make(chan func()([]Action, func())),
		roomID: roomID,
		state: Waiting,
		ready: make(map[string]bool),
		quit: make(chan struct{}),
		musicProvider: musicProvider,
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