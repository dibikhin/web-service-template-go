package domain

type NotFoundError struct {
	Message string
}

func NewNotFoundError(msg string) error {
	return &NotFoundError{
		Message: msg,
	}
}

func (e *NotFoundError) Error() string {
	return e.Message
}
