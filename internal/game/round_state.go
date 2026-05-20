package game

type RoundState int

const (
	Idle RoundState = iota
	Loading
	Playing
	Result
	Finished
)

func (s RoundState) String() string {
	switch s {
		case Idle:
			return "idle"
		case Loading:
			return "loading"
		case Playing:
			return "playing"
		case Result:
			return "result"
		case Finished:
			return "finished"
	}

	return "unknown"
}