package message

import (
	"temp-ws/internal/music"
)

type WelcomePayload struct {
	HostID string		`json:"hostID"`
	PlayerID string		`json:"playerID"`
	RoomID string		`json:"roomID"`
	Players []string	`json:"players"`
	RoomState string	`json:"roomState"`
	MaxPlayers int		`json:"maxPlayers"`
}

type PlayerJoinedPayload struct {
	PlayerID string			`json:"playerID"`
}

type PlayerLeftPayload struct {
	PlayerID string			`json:"playerID"`
}

type HostChangedPayload struct {
	PlayerID string			`json:"playerID"`
}

type StartGamePayload struct {
	Category music.Category	`json:"category"`
	TrackCount int			`json:"trackCount"`
	TimeLimit int			`json:"timeLimit"`
}

type ReadyToPlayPayload struct {
    RoundNumber int			`json:"roundNumber"`
}

type CountDownPayload struct {
	Remaining int	`json:"remaining"`
}

type KickPlayerPayload struct {
	TargetPlayerID string		`json:"targetPlayerID"`
}

type SubmitAnswerPayload struct {
	Answer string		`json:"answer"`
}

type PreloadSongPayload struct {
	ISRC string		`json:"isrc"`
	StartTime int	`json:"startTime"`
	RoundNumber int    `json:"roundNumber"`
}

type RoundStartPayload struct {
	RoundNumber int 		`json:"roundNumber"`
	TotalRounds int			`json:"totalRounds"`
	AnswerLength int		`json:"answerLength"`
	LetterCards []string	`json:"letterCards"`
}

type RoundResultPayload struct {
	Winner string 			`json:"winner"`
	CorrectAnswer string	`json:"correctAnswer"`
	Scores map[string]int 	`json:"scores"`
}

type WrongAnswerPayload struct {
	PlayerID string			`json:"playerID"`
	WrongAnswer string		`json:"wrongAnswer"`
}

type GameOverPayload struct {
	Winner string			`json:"winner"`
	Scores map[string]int	`json:"scores"`
}