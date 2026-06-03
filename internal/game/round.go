package game

import "temp-ws/internal/music"

type Round struct {
	Winner string
	Answer string
	LetterCards []string
	Song music.Song
	State RoundState
	StartTime int
 }