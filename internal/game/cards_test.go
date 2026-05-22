package game

import (
	"slices"
	"testing"
	"unicode"
)

func TestExtractCard(t *testing.T) {
	title := "어떻게 이별까지 사랑하겠어, 널 사랑하는 거지"
	expected := []rune { '어', '떻', '게', '이', '별', '까', '지', '사', '랑', '하', '겠', '어', '널', '사', '랑', '하', '는', '거', '지'  }

	cards := extractCards(title)

	if !slices.Equal(cards, expected) {
		t.Errorf("제목 카드 추출 실패\n got: %c\nwant: %c", cards, expected)
	}
}

func TestDecideTotalCardCount(t *testing.T) {
	n := decideTotalCardCount(17)

	if n != 25 {
		t.Errorf("정답 + 더미 카드 총 개수 다름\n got: %d\nwant: %d", n, 25)
	}
}

func TestPickDummiesOnlyHangul(t *testing.T) {
	answer := []rune { '뜨', '거', '운', '여', '름', '밤', '은', '가', '고', '남', '은', '건', '볼', '품', '없', '지', '만',  }

	dummies := pickDummies(answer, 8)

	if len(dummies) != 8 {
		t.Error("더미 개수 다름")
	}

	answerMap := map[rune]bool{}
	for _, r := range answer {
		answerMap[r] = true
	}

	for _, r := range dummies {
		if answerMap[r] {
			t.Error("더미와 정답이 겹침")
		}

		if !unicode.Is(unicode.Hangul, r) {
			t.Error("한글 정답인데 한글이 아닌 더미가 나옴")
		}
	}
}

func TestPickDummiesHangulAndNumber(t *testing.T) {
	answer := []rune { '비', '밀', '번', '호', '4', '8', '6' }
	
	dummies := pickDummies(answer, 3)

	if len(dummies) != 3 {
		t.Error("더미 개수 다름")
	}

	answerMap := map[rune]bool{}
	for _, r := range answer {
		answerMap[r] = true
	}

	for _, r := range dummies {
		if answerMap[r] {
			t.Error("더미와 정답이 겹침")
		}

		if r >= 'A' && r <= 'Z' {
			t.Error("정답에 영어가 없는데 영어 더미 나옴")
		}
	}

	var hangulCount, digitCount int

	for _, r := range dummies {
		switch {
			case unicode.Is(unicode.Hangul, r):
				hangulCount += 1
			case unicode.IsDigit(r):
				digitCount += 1
		}
	}

	if hangulCount != 2 || digitCount != 1 {
		t.Error("비율 배분 틀림")
	}
}

func TestPickDummiesOnlyEnglish(t *testing.T) {
	answer := []rune { 'I', 'B', 'E', 'L', 'I', 'E', 'I', 'V', 'E' }

	dummies := pickDummies(answer, 13)

	if len(dummies) != 13 {
		t.Error("더미 개수 다름")
	}

	answerMap := map[rune]bool{}
	for _, r := range answer {
		answerMap[r] = true
	}

	for _, r := range dummies {
		if answerMap[r] {
			t.Error("더미와 정답이 겹침")
		}

		if r < 'A' || r > 'Z' {
			t.Error("정답에 영어밖에 없는데 다른 더미 나옴")
		}
	}
}

func TestDistributeByRatio(t *testing.T) {
	dist := distributeByRatio(6, []int { 2, 2, 2 })

	if !slices.Equal(dist, []int { 2, 2, 2 }) {
		t.Error("더미 분배 로직 틀림")
	}

	dist = distributeByRatio(5, []int { 3, 1, 1 })

	if !slices.Equal(dist, []int { 3, 1, 1 }) {
		t.Error("더미 분배 로직 틀림")
	}

	dist = distributeByRatio(5, []int { 2, 0, 1 })

	if !slices.Equal(dist, []int { 3, 0, 2 }) {
		t.Error("더미 분배 로직 틀림")
	}

	dist = distributeByRatio(5, []int { 0, 0, 0 })

	if !slices.Equal(dist, []int { 0, 0, 0 }) {
		t.Error("더미 분배 로직 틀림")
	}
}