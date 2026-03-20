package score

import "time"

type ScoreEntry struct {
	BoardID    string
	PeriodID   int64
	UserID     string
	Score      int64
	AchievedAt time.Time
}

type RankedEntry struct {
	ScoreEntry
	Rank int
}
