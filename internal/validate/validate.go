package validate

import (
	"errors"
	"strings"
)

func BoardID(boardID string) error {
	if strings.TrimSpace(boardID) == "" {
		return errors.New("board ID must not be empty")
	}

	return nil
}
