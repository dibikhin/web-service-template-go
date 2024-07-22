package dummy

import "errors"

var (
	ErrNotFound = errors.New("not found")
)

type NotFoundError struct {
	Err     error
	Message string
}

func NewNotFoundError(msg string) error {
	return &NotFoundError{
		Message: msg, Err: ErrNotFound,
	}
}

func (e *NotFoundError) Error() string {
	return e.Message
}

func (e *NotFoundError) Is(target error) bool {
	return errors.Is(e.Err, target)
}
