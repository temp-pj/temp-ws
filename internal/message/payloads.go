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
}

type KickPlayerPayload struct {
	TargetPlayerID string
}

type SubmitAnswerPayload struct {
	Answer string
}

type PreloadSongPayload struct {
	URL string `json:"url"`
}

type RoundStartPayload struct {
	LetterCards []string	`json:"letterCards"`
	TimeLimit int		`json:"timeLimit"`
}

type RoundResultPayload struct {
	Winner string 			`json:"winner"`
	Scores map[string]int 	`json:"scores"`
}

type WrongAnswerPayload struct {
	WrongAnswer string		`json:"wrongAnswer"`
}

type GameOverPayload struct {
	Winner string			`json:"winner"`
	Score map[string]int	`json:"score"`
}