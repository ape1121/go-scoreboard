package board

import "time"

type ScheduleType string

const (
	ScheduleTypeInterval ScheduleType = "interval"
)

type Board struct {
	ID          string
	Name        string
	Description string
	Schedule    *Schedule
	CreatedAt   time.Time
}

func (b Board) HasSchedule() bool {
	return b.Schedule != nil
}

type Schedule struct {
	Type     ScheduleType
	Interval time.Duration
}

func (s Schedule) IntervalSeconds() int64 {
	return int64(s.Interval / time.Second)
}

type BoardPeriod struct {
	ID        int64
	BoardID   string
	Sequence  int64
	StartedAt time.Time
	EndedAt   *time.Time
}

func (p BoardPeriod) IsActive() bool {
	return p.EndedAt == nil
}
