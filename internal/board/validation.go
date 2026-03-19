package board

import (
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	MinNameLength        = 1
	MaxNameLength        = 100
	MaxDescriptionLength = 500
	MinScheduleInterval  = time.Second
	MaxScheduleInterval  = 365 * 24 * time.Hour
)

func ValidateName(name string) error {
	trimmed := strings.TrimSpace(name)
	length := utf8.RuneCountInString(trimmed)

	switch {
	case length < MinNameLength:
		return fmt.Errorf("name must not be empty")
	case length > MaxNameLength:
		return fmt.Errorf("name must be at most %d characters", MaxNameLength)
	default:
		return nil
	}
}

func ValidateDescription(description string) error {
	if utf8.RuneCountInString(description) > MaxDescriptionLength {
		return fmt.Errorf("description must be at most %d characters", MaxDescriptionLength)
	}

	return nil
}

func ValidateSchedule(schedule *Schedule) error {
	if schedule == nil {
		return nil
	}

	if schedule.Type != ScheduleTypeInterval {
		return fmt.Errorf("schedule type must be %q", ScheduleTypeInterval)
	}

	if schedule.Interval < MinScheduleInterval {
		return fmt.Errorf("schedule interval must be at least %s", MinScheduleInterval)
	}

	if schedule.Interval > MaxScheduleInterval {
		return fmt.Errorf("schedule interval must be at most %s", MaxScheduleInterval)
	}

	if schedule.Interval%time.Second != 0 {
		return fmt.Errorf("schedule interval must be expressed in whole seconds")
	}

	return nil
}

func ValidateNewBoard(name, description string, schedule *Schedule) error {
	return errors.Join(
		ValidateName(name),
		ValidateDescription(description),
		ValidateSchedule(schedule),
	)
}
