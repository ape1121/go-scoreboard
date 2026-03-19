package board

import "time"

type CreateInput struct {
	Name        string
	Description string
	Schedule    *Schedule
}

type Details struct {
	Board       Board
	NextResetAt *time.Time
}
