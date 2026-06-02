package game

import (
	"math"
	"temp-ws/internal/music"
	"time"
)

type Game struct {
	rounds []Round
	currentRound int
	scores map[string]int
	roundDuration time.Duration
	roundStartedAt time.Time
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

	 return RoundStartInfo { RoundNumber: g.currentRound + 1, TotalRounds: len(g.rounds), LetterCards: r.LetterCards  }
}

func (g *Game) CurrentISRC()string {
	return g.rounds[g.currentRound].Song.ISRC
}

func (g *Game) IsPlaying()bool {
	return g.rounds[g.currentRound].State == Playing
}

func (g *Game) StartRound(onTimeout func()) {
	g.rounds[g.currentRound].State = Playing
	g.roundStartedAt = time.Now()
	g.roundTimer = time.AfterFunc(g.roundDuration, onTimeout)
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

func (g *Game) RemainingTime()int {
	elapsed := time.Since(g.roundStartedAt)
	remainingTime := g.roundDuration - elapsed

	if remainingTime <= 0 { return 0 }

	return int(math.Ceil(remainingTime.Seconds()))
}

func (g *Game) StopTimer() {
	if g.roundTimer != nil {
		g.roundTimer.Stop()
	}
}

func NewGame(playerIDs []string, songs []music.Song, timeLimit int) *Game {
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

	roundDuration := time.Duration(timeLimit) * time.Second

	return &Game { rounds: rounds, currentRound: 0, scores: scores, roundDuration: roundDuration }
}