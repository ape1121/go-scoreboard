package score

import "time"

type SetInput struct {
	BoardID string
	UserID  string
	Score   int64
}

type UpsertInput struct {
	BoardID    string
	UserID     string
	Score      int64
	AchievedAt time.Time
}
