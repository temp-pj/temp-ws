package room

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"temp-ws/internal/client"
	"temp-ws/internal/game"
	"temp-ws/internal/message"
	"temp-ws/internal/music"
	"time"
)

type MusicProvider interface {
	FetchSongs(ctx context.Context, category music.Category, count int, timeLimit int) ([]music.Song, error)
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
	onEmpty func(roomID string)
}

func (r *Room) Run() {
	ticker := time.NewTicker(1 * time.Second)

	defer ticker.Stop()

	for {
		select {
			case client := <- r.register:
				if client == nil || client.Send == nil { continue }

				if r.hostID == "" { r.hostID = client.ID }
				r.clients[client.ID] = client

				playerIDs := make([]string, 0, len(r.clients))
				for id := range r.clients {
					playerIDs = append(playerIDs, id)
				}

				welcomeMsg, err := message.New(message.TypeWelcome, message.WelcomePayload { HostID: r.hostID, 
					RoomID: r.roomID, 
					PlayerID: client.ID, 
					Players: playerIDs,
					RoomState: r.state.String(),
					MaxPlayers: 8,
				  })
				if err != nil { continue }

				playerJoinedMsg, err := message.New(message.TypePlayerJoined, message.PlayerJoinedPayload { PlayerID: client.ID })
				if err != nil { continue }

				select {
					case client.Send <- welcomeMsg:

					default:
						r.removeClient(client.ID)
						continue
				}
				
				var dead []string

				for _, c := range r.clients {
					if c.ID == client.ID { continue }

					select {
						case c.Send <- playerJoinedMsg:

						default:
							dead = append(dead, c.ID)
					}
				}

				for _, id := range dead {
					r.removeClient(id)
				}

			case client := <- r.unregister:
				hostChanged := r.removeClient(client.ID)

				msg, err := message.New(message.TypePlayerLeft, message.PlayerLeftPayload { PlayerID: client.ID })
				if err != nil { continue }

				var dead []string
				for _, c := range r.clients {
					select {
						case c.Send <- msg:

						default:
							dead = append(dead, c.ID)
					}
				}

				for _, id := range dead { r.removeClient(id) }

				if hostChanged && r.hostID != "" {
					hostMsg, err := message.New(message.TypeHostChanged, message.HostChangedPayload{ PlayerID: r.hostID })
					if err == nil {
						var dead2 []string
						for _, c := range r.clients {
							select {
								case c.Send <- hostMsg:
								default:
									dead2 = append(dead2, c.ID)
							}
						}
						for _, id := range dead2 { r.removeClient(id) }
					}
				}

				if len(r.clients) == 0 {
					if r.onEmpty != nil { go r.onEmpty(r.roomID) }
					continue
				}
			
			case msg := <- r.broadcast:
				var dead []string

				for _, c := range r.clients {
					select {
						case c.Send <- msg:

						default: 
							dead = append(dead, c.ID)
					}
				}

				for _, id := range dead { r.removeClient(id) }

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
				if !r.game.IsPlaying() { continue }

				remaining := r.game.RemainingTime()
				msg, err := message.New(message.TypeCountDown, message.CountDownPayload { Remaining: remaining })
				if err != nil { continue }

				var dead []string
				for _, c := range r.clients {
					select {
					case c.Send <- msg:
					default:
						dead = append(dead, c.ID)
					}
				}

				for _, id := range dead {
					r.removeClient(id)
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
			log.Println("HOST ID: ", r.hostID)
			if msg.From != r.hostID { return nil, nil }
			var payload message.StartGamePayload

			err := json.Unmarshal(msg.Message.Payload, &payload)
			log.Println("Unmarshal 결과:", err, payload)
			if err != nil { return nil, nil }

			playerIDs := make([]string, 0, len(r.clients))
			for id := range r.clients {
				playerIDs = append(playerIDs, id)
			}
			
			category := payload.Category
			trackCount := payload.TrackCount

			if trackCount <= 0 { trackCount = 50 }
			if payload.TimeLimit <= 0 { payload.TimeLimit = 30 }
			
			go func() {
				songs, err := r.musicProvider.FetchSongs(context.Background(), category, trackCount, payload.TimeLimit )
				
				log.Println("FetchSongs 결과:", len(songs), err)

				if err != nil || len(songs) == 0 { return }
				
				r.internal <- func() ([]Action, func()) {
					r.game = game.NewGame(playerIDs, songs, payload.TimeLimit)
					isrc := r.game.CurrentISRC()
					startTime := r.game.CurrentStartTime()
					roundNumber := r.game.CurrentRound()
					preloadSongPayload := message.PreloadSongPayload { ISRC: isrc, StartTime: startTime, RoundNumber: roundNumber }
					preloadSongMsg, _ := message.New(message.TypePreloadSong, preloadSongPayload)

					return []Action {
						{ Type: Broadcast, Message: preloadSongMsg },
					}, nil
				}
			}()
			
			r.state = Playing
			gameStartedMsg, _ := message.New(message.TypeGameStarted, nil)

			return []Action{ {Type: Broadcast, Message: gameStartedMsg }, }, nil

		case "READY_TO_PLAY":
			log.Println("READY_TO_PLAY 시작")
			log.Println("READY_TO_PLAY 요청한 ID: ", msg.From)

			if r.game == nil { return nil, nil }
			log.Println("게임 있음")

			var payload message.ReadyToPlayPayload
			err := json.Unmarshal(msg.Message.Payload, &payload)

			if err != nil { return nil, nil }
			log.Println("JSON 언마샬")

			if payload.RoundNumber != r.game.CurrentRound() { 
				log.Printf("페이로드 라운드 넘버: %d, 서버 라운드 넘버: %d", payload.RoundNumber, r.game.CurrentRound())
				return nil, nil 
			}
			log.Println("라운드 넘버 통과")

			r.ready[msg.From] = true
			allReady := true

			log.Println("=== clients ===")

			for id := range r.clients {
				log.Println("READY ID: ", id)
				if !r.ready[id] {  
					allReady = false
					log.Printf("client=%s ready=%v\n", id, r.ready[id])
					break
				}
			}

			log.Printf("총 clients=%d, ready맵=%v\n", len(r.clients), r.ready)

			if !allReady { return nil, nil }

			log.Println("전부 준비 완료")

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

			log.Println("서버 라운드 스타트")
			
			roundStartInfo := r.game.GetRoundStartInfo()

			roundStartedPayload := message.RoundStartPayload { 
				RoundNumber: roundStartInfo.RoundNumber, 
				TotalRounds: roundStartInfo.TotalRounds, 
				LetterCards: roundStartInfo.LetterCards,
				AnswerLength: roundStartInfo.AnswerLength,
			}

			roundStartMsg, _ := message.New(message.TypeRoundStarted, roundStartedPayload)
			log.Println("브로드캐스트까지 성공")
			return []Action { { Type: Broadcast, Message: roundStartMsg } }, nil


		case "KICK_PLAYER":
			if msg.From != r.hostID { return nil, nil }
			
			var payload message.KickPlayerPayload

			err := json.Unmarshal(msg.Message.Payload, &payload)
			if err != nil { return nil, nil }

			target := payload.TargetPlayerID

			if _, ok := r.clients[target]; !ok { return nil, nil }

			kickPlayerMsg, _ := message.New(message.TypePlayerKicked, nil)
		
			return []Action { { Type: Unicast, Target: target, Message: kickPlayerMsg } }, func() { 
					r.removeClient(target)
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
				wrongAnswerMsg, _ := message.New(message.TypeWrongAnswer, wrongAnswerPayload)

				return []Action{
					{Type: Broadcast, Message: wrongAnswerMsg },
				}, nil
			}

			return r.endRoundActions(msg.From)
	}

	return nil, nil
}

func (r *Room) removeClient(id string) bool {
	c, ok := r.clients[id]

	if !ok { return false }

	close(c.Send)
	delete(r.clients, id)
	delete(r.ready, id)


	if id == r.hostID {
        r.hostID = ""
        for remainingID := range r.clients {
            r.hostID = remainingID
            break
        }

		return true
    }

	return false
}

func (r *Room) endRoundActions(winner string) ([]Action, func()) {
    r.game.EndRound(winner)

    roundResultMsg, _ := message.New(message.TypeRoundResult, message.RoundResultPayload{
        Winner: winner, CorrectAnswer: r.game.CurrentAnswer(), Scores: r.game.Scores(),
    })

    if r.game.NextRound() {
		isrc := r.game.CurrentISRC()
		startTime := r.game.CurrentStartTime()
        preloadMsg, _ := message.New(message.TypePreloadSong, message.PreloadSongPayload{
            ISRC: isrc,
			StartTime: startTime,
			RoundNumber: r.game.CurrentRound(),
        })

        return []Action{
            {Type: Broadcast, Message: roundResultMsg},
            {Type: Broadcast, Message: preloadMsg},
        }, nil
    }

    gameOverMsg, _ := message.New(message.TypeGameOver, message.GameOverPayload{
        Winner: r.game.Winner(),
		Scores: r.game.Scores(),
    })

	r.state = Result

    return []Action{
		{Type: Broadcast, Message: roundResultMsg},
        {Type: Broadcast, Message: gameOverMsg},
    }, func() { r.game = nil }
}

func (r *Room) executeActions(actions []Action) {
	for _, a := range actions {
		switch a.Type {
			case Broadcast:
				var dead []string
				log.Printf("브로드캐스트: %d명에게", len(r.clients))
				for id, c := range r.clients {
					select {
					case c.Send <- a.Message:
					default:
						dead = append(dead, id)
					}
				}

				for _, id := range dead { r.removeClient(id) }

			case Unicast:
				if target, ok := r.clients[a.Target]; ok {
					select {
						case target.Send <- a.Message:
						default:
							r.removeClient(a.Target)
					}
				}
		}
	}
}

func (r *Room) IsPlaying()bool {
	return r.state == Playing
}

func NewRoom(roomID string, musicProvider MusicProvider, onEmpty func(string)) *Room {
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
		onEmpty: onEmpty,
	}
}


type RoomState int

const (
	Waiting RoomState = iota
	Playing
	Result
)

func (s RoomState) String() string {
	switch s {
		case Waiting:
			return "waiting"
		case Playing:
			return "playing"
		case Result:
			return "result"
	}

	return "unknown"
}