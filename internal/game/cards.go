package game

import (
	"math/rand"
	"sort"
	"unicode"
)

func generateLetterCards(title string) []string {
	answer := extractCards(title)

	if len(answer) == 0 { return nil }

	total := decideTotalCardCount(len(answer))
	dummyCount := total - len(answer)
	dummies := pickDummies(answer, dummyCount)

	all := append(answer, dummies...)

	rand.Shuffle(len(all), func (i, j int) {
		all[i], all[j] = all[j], all[i]
	})
	
	return runesToString(all)
}

func extractCards(title string) []rune {
	var cards []rune

	for _, r := range title {
		switch {
			case unicode.Is(unicode.Hangul, r):
				cards = append(cards, r)

			case r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z':
				cards = append(cards, unicode.ToUpper(r))

			case unicode.IsDigit(r):
				cards = append(cards, r)
		}
	}

	return cards
}

func decideTotalCardCount(n int) int {
	total := n + n/2
	if total < 8 {
		return 8
	}

	return total
}

func pickDummies(answer []rune, count int) []rune {
	var koreanCount int
	var englishCount int
	var digitCount int
	
	for _, r := range answer {
		switch {
			case unicode.Is(unicode.Hangul, r):
				koreanCount += 1
			case r >= 'A' && r <= 'Z':
				englishCount += 1
			case unicode.IsDigit(r):
				digitCount += 1
		}
	}

	total := koreanCount + englishCount + digitCount

	if total == 0 || count == 0 { return nil }

	dist := distributeByRatio(count, []int { koreanCount, englishCount, digitCount })

	var dummies []rune
	
	dummies = append(dummies, pickFromPool(koreanPool, answer, dist[0])...)
	dummies = append(dummies, pickFromPool(englishPool, answer, dist[1])...)
	dummies = append(dummies, pickFromPool(digitPool, answer, dist[2])...)

	return dummies
}

func distributeByRatio(total int, weights []int) []int {
    sum := 0
    for _, w := range weights {
        sum += w
    }
    if sum == 0 {
        return make([]int, len(weights))
    }

    ratios := make([]float64, len(weights))
    result := make([]int, len(weights))
    remain := total

    for i, w := range weights {
        ratios[i] = float64(total) * float64(w) / float64(sum)
        result[i] = int(ratios[i])
        remain -= result[i]
    }

    order := make([]int, len(weights))
    for i := range order {
        order[i] = i
    }
    sort.Slice(order, func(i, j int) bool {
        return ratios[order[i]]-float64(result[order[i]]) > ratios[order[j]]-float64(result[order[j]])
    })

    for i := 0; remain > 0 && i < len(order); i++ {
        result[order[i]] += 1
        remain -= 1
    }

    return result
}

func pickFromPool(pool []rune, answer []rune, count int) []rune {
	answerMap := map[rune]bool { }
	pickedMap := map[rune]int { }
	dummies := []rune {}

	for _, c := range answer {
		answerMap[c] = true
	}

	i := 0
	attempts := 0

	for i < count {
		if attempts > count * 10 {
			break
		}
		attempts++

		pick := pool[rand.Intn(len(pool))]
		if answerMap[pick] { continue }
		if pickedMap[pick] > 1 { continue }

		dummies = append(dummies, pick)
		pickedMap[pick] += 1
		i += 1
	}

	return dummies
}

func runesToString(runes []rune) []string {
	result := make([]string, len(runes))
    for i, r := range runes {
        result[i] = string(r)
    }
    return result
}