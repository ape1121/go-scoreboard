package score

import "errors"

var (
	ErrBoardNotFound        = errors.New("board not found")
	ErrActivePeriodNotFound = errors.New("active period not found")
	ErrScoreNotFound        = errors.New("score not found")
)

type ValidationError struct {
	err error
}

func NewValidationError(err error) error {
	if err == nil {
		return nil
	}

	return ValidationError{err: err}
}

func (e ValidationError) Error() string {
	return e.err.Error()
}

func (e ValidationError) Unwrap() error {
	return e.err
}
