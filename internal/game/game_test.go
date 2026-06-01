package game

import (
	"temp-ws/internal/music"
	"testing"

	"github.com/google/uuid"
)


func TestFullGameCycle(t *testing.T) {
	player1 := uuid.NewString()
	player2 := uuid.NewString()
	player3 := uuid.NewString()

	playerID := []string { player1, player2, player3 }
	songs := []music.Song { 
		{ Title: "외딴섬 로맨틱", Artist: "잔나비", }, 
		{ Title: "초록을거머쥔우리는", Artist: "잔나비", }, 
		{ Title: "주저하는연인들을위해", Artist: "잔나비", }, 
		{ Title: "꿈과 책과 힘과 벽", Artist: "잔나비", }, 
 }
	game := NewGame(playerID, songs)

	game.StartRound(func() { })

	correct := game.SubmitAnswer("외딴섬 로맨틱")

	if !correct {
		t.Fatal("라운드1 정답 판정 실패")
	}
	game.EndRound(player1)

	if !game.NextRound() {
		t.Fatal("라운드2로 넘어가야 하는데 false 반환")
	}

	game.StartRound(func() { })

	wrong := game.SubmitAnswer("사랑하긴했었나요스쳐가는인연이었나요")

	if wrong {
		t.Fatal("오답인데 true 반환")
	}

	correct = game.SubmitAnswer("초록을거머쥔우리는")

	if !correct {
		t.Fatal("라운드2 정답 판정 실패")
	}
	game.EndRound(player2)

	if !game.NextRound() {
		t.Fatal("라운드3으로 넘어가야 하는데 false 반환")
	}

	game.StartRound(func() { })

	correct = game.SubmitAnswer("주저하는연인들을위해")

	if !correct {
		t.Fatal("라운드3 정답 판정 실패")
	}
	
	game.EndRound(player3)

	if !game.NextRound() {
		t.Fatal("라운드3으로 넘어가야 하는데 false 반환")
	}

	game.StartRound(func() { })

	game.EndRound("")

	if game.NextRound() {
		t.Fatal("마지막 라운드인지 모르고 true 반환")
	}

	scores := game.Scores()

	for _, id := range playerID {
		if scores[id] != 1 {
			t.Error("점수 반영 이상")
		}
	}
}

func TestNewGame(t *testing.T) {
	playerID := []string { uuid.NewString(), uuid.NewString(), uuid.NewString()  }
	songs := []music.Song { 
		{ Title: "외딴섬 로맨틱", Artist: "잔나비", }, 
		{ Title: "초록을거머쥔우리는", Artist: "잔나비", }, 
		{ Title: "주저하는연인들을위해", Artist: "잔나비", }, 
 }
	game := NewGame(playerID, songs)

	scores := game.Scores()
	round := game.CurrentRound()

	for _, id := range playerID {
		if scores[id] != 0 {
			t.Errorf("플레이어 %s 초기 점수가 0이 아님: %d", id, scores[id])
		}
	}
	if round != 1 { t.Error("게임 생성 시 round 값 이상") }
}

func TestStartRound(t *testing.T) {
	playerID := []string { uuid.NewString(), uuid.NewString(), uuid.NewString()  }
	songs := []music.Song { 
		{ Title: "외딴섬 로맨틱", Artist: "잔나비", }, 
		{ Title: "초록을거머쥔우리는", Artist: "잔나비", }, 
		{ Title: "주저하는연인들을위해", Artist: "잔나비", }, 
 	}
	game := NewGame(playerID, songs)

	game.StartRound(func() { })
	round := game.rounds[game.currentRound]

	if round.State != Playing {
		t.Error("Round 상태 이상")
	}

	defer game.StopTimer()
}

func TestEndRound_WithWinner(t *testing.T) {
	playerID := []string { uuid.NewString(), uuid.NewString(), uuid.NewString()  }
	songs := []music.Song { 
		{ Title: "외딴섬 로맨틱", Artist: "잔나비", }, 
		{ Title: "초록을거머쥔우리는", Artist: "잔나비", }, 
		{ Title: "주저하는연인들을위해", Artist: "잔나비", }, 
 	}
	game := NewGame(playerID, songs)

	game.StartRound(func() { })

	correct := game.SubmitAnswer("외딴섬 로맨틱")

	if !correct {
		t.Error("답 비교 로직 문제있음")
	}

	game.EndRound(playerID[0])
	scores := game.Scores()
	round := game.rounds[game.currentRound]

	if scores[playerID[0]] != 1 { 
		t.Error("정답을 맞춘 플레이어 점수 반영 안됨")
	 }

	if round.State != Result {
		t.Error("라운드 상태 변경 안됨")
	}

	if round.Winner != playerID[0] {
		t.Error("정답자 아이디 반영 안됨")
	}
}

func TestNextRound_Timeout(t *testing.T) {
	playerID := []string { uuid.NewString(), uuid.NewString(), uuid.NewString()  }
	songs := []music.Song { 
		{ Title: "외딴섬 로맨틱", Artist: "잔나비", }, 
		{ Title: "초록을거머쥔우리는", Artist: "잔나비", }, 
		{ Title: "주저하는연인들을위해", Artist: "잔나비", }, 
 	}
	game := NewGame(playerID, songs)

	game.StartRound(func() { })

	game.EndRound("")
	scores := game.Scores()

	for _, id := range playerID {
		if scores[id] != 0 {
			t.Error("시간초과로 인한 라운드 종료인데 점수 변화 생김")
		}
	}

	round := game.rounds[game.currentRound]

	if round.State != Result {
		t.Error("라운드 상태 변경 안됨")
	}

	if round.Winner != "" {
		t.Error("정답자 없음 표기 안됨")
	}
}

func TestSubmitAnswer(t *testing.T) {
	playerID := []string { uuid.NewString(), uuid.NewString(), uuid.NewString()  }
	songs := []music.Song { 
		{ Title: "외딴섬 로맨틱", Artist: "잔나비", }, 
		{ Title: "초록을거머쥔우리는", Artist: "잔나비", }, 
		{ Title: "주저하는연인들을위해", Artist: "잔나비", }, 
 	}
	game := NewGame(playerID, songs)

	game.StartRound(func() { })

	correct := game.SubmitAnswer("외딴섬 로맨틱")

	if !correct {
		t.Error("답 비교 로직 문제있음")
	}

	wrong := game.SubmitAnswer("바보")

	if wrong {
		t.Error("오답인데 정답 됨")
	}
}

