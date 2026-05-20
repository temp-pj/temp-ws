package game

type RoundState int

const (
	Idle RoundState = iota
	Playing
	Result
)

func (s RoundState) String() string {
	switch s {
		case Idle:
			return "idle"
		case Playing:
			return "playing"
		case Result:
			return "result"
	}

	return "unknown"
}