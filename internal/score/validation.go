package score

import (
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"
)

const (
	MinScore        int64 = 0
	MaxQueryLimit         = 100
	MinQueryLimit         = 1
	MaxUserIDLength       = 128
)

func ValidateUserID(userID string) error {
	trimmed := strings.TrimSpace(userID)
	length := utf8.RuneCountInString(trimmed)

	switch {
	case length == 0:
		return fmt.Errorf("user ID must not be empty")
	case length > MaxUserIDLength:
		return fmt.Errorf("user ID must be at most %d characters", MaxUserIDLength)
	default:
		return nil
	}
}

func ValidateValue(value int64) error {
	if value < MinScore {
		return fmt.Errorf("score must be greater than or equal to %d", MinScore)
	}

	return nil
}

func ValidateLimit(limit int) error {
	switch {
	case limit < MinQueryLimit:
		return fmt.Errorf("limit must be at least %d", MinQueryLimit)
	case limit > MaxQueryLimit:
		return fmt.Errorf("limit must be at most %d", MaxQueryLimit)
	default:
		return nil
	}
}

func ValidateWrite(userID string, value int64) error {
	return errors.Join(
		ValidateUserID(userID),
		ValidateValue(value),
	)
}
