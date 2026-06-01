package message

import (
	"temp-ws/internal/music"
)

type PlayerJoinedPayload struct {
	PlayerID string 
}

type PlayerLeftPayload struct {
	PlayerID string 
}

type StartGamePayload struct {
	Category music.Category
	TrackCount int
}

type ReadyToPlayPayload struct {
    RoundNumber int
}

type KickPlayerPayload struct {
	TargetPlayerID string
}

type SubmitAnswerPayload struct {
	Answer string
}

type PreloadSongPayload struct {
	ISRC string		`json:"isrc"`
}

type RoundStartPayload struct {
	RoundNumber int 		`json:"roundNumber"`
	TotalRound int			`json:"totalRound"`
	LetterCards []string	`json:"letterCards"`
	TimeLimit int			`json:"timeLimit"`
}

type RoundResultPayload struct {
	Winner string 			`json:"winner"`
	Scores map[string]int 	`json:"scores"`
}

type WrongAnswerPayload struct {
	PlayerID string			`json:"playerID"`
	WrongAnswer string		`json:"wrongAnswer"`
}

type GameOverPayload struct {
	Winner string			`json:"winner"`
	Score map[string]int	`json:"score"`
}