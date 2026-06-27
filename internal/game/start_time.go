package game

import "math/rand/v2"

func generateStartTime(durationSec int, timeLimitSec int) int {
	maxSec := durationSec - 10 - timeLimitSec

	if maxSec <= 0 { return 0 }

	return rand.IntN(maxSec+1)
}