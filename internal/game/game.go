package game

import (
	"temp-ws/internal/music"
	"time"
)

type Game struct {
	rounds []Round
	currentRound int
	scores map[string]int
	roundTimer *time.Timer
}

func (g *Game) Scores()map[string]int {
	copied := make(map[string]int, len(g.scores))
    for k, v := range g.scores {
        copied[k] = v
    }
    return copied
}

func (g *Game) CurrentRound()int {
	return g.currentRound + 1
}

func (g *Game) GetRoundStartInfo() RoundStartInfo {
	 r := g.rounds[g.currentRound]

	 return RoundStartInfo { RoundNumber: g.currentRound + 1, TotalRounds: len(g.rounds), LetterCards: r.LetterCards, TimeLimit: 30  }
}

func (g *Game) CurrentISRC()string {
	return g.rounds[g.currentRound].Song.ISRC
}

func (g *Game) IsPlaying()bool {
	return g.rounds[g.currentRound].State == Playing
}

func (g *Game) StartRound(onTimeout func()) {
	g.rounds[g.currentRound].State = Playing
	g.roundTimer = time.AfterFunc(30 * time.Second, onTimeout)
}

func (g *Game) EndRound(winnerID string) {
	if g.roundTimer != nil {
		g.roundTimer.Stop()
	}

	g.rounds[g.currentRound].State = Result
	g.rounds[g.currentRound].Winner = winnerID
	
	if winnerID != "" {
		g.scores[winnerID] += 1
	}
}

func (g *Game) NextRound()bool {
	if len(g.rounds) <= g.currentRound+1 { return false }

	g.currentRound += 1
	return true
}

func (g *Game) SubmitAnswer(answer string)bool {
	return answer == g.rounds[g.currentRound].Answer
}

func (g *Game) StopTimer() {
	if g.roundTimer != nil {
		g.roundTimer.Stop()
	}
}

func NewGame(playerIDs []string, songs []music.Song) *Game {
	scores := make(map[string]int)

	for _, id := range playerIDs {
		scores[id] = 0
	}

	rounds := make([]Round, len(songs))

	for i, song := range songs {
		rounds[i] = Round {
			Answer: song.Title, 
			LetterCards: generateLetterCards(song.Title),
			Song: song,
			State: Idle,
		}
	}

	return &Game { rounds: rounds, currentRound: 0, scores: scores }
}