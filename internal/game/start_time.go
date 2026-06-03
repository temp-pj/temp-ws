package game

import "math/rand/v2"

func generateStartTime(durationMs int, timeLimitSec int)int {
	durationSec := durationMs / 1000
	maxSec := durationSec - 10 - timeLimitSec

	if maxSec <= 0 { return 0 }

	return rand.IntN(maxSec)
}