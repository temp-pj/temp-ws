package game

import (
	"testing"
)

func TestGenerateStartTime(t *testing.T) {
	t.Run("유효 범위 내 startTime 생성", func(t *testing.T) {
		startTime := generateStartTime(180000, 60)
		if startTime < 0 || startTime >= 110 {
			t.Errorf("범위 밖: %d", startTime)
		}
	})

	t.Run("duration이 최소 기준과 같을 때 panic 안 남", func(t *testing.T) {
		startTime := generateStartTime(70000, 60)
		if startTime != 0 {
			t.Errorf("startTime은 0이어야 함: %d", startTime)
		}
	})
}