package middleware

import (
	"encoding/json"
	"net/http"
)

type (
	ErrorResponse struct {
		Error APIError `json:"error"`
	}

	APIError struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
)

// 400 Bad Request

type ValidationError struct {
	Message string
}

func NewValidationError(msg string) error {
	return &ValidationError{Message: msg}
}

func (e *ValidationError) Error() string {
	return e.Message
}

func (ValidationError) StatusCode() int {
	return http.StatusBadRequest
}

func (e *ValidationError) MarshalJSON() ([]byte, error) {
	return json.Marshal(&ErrorResponse{
		Error: APIError{
			Code:    60801,
			Message: e.Error(),
		},
	})
}

// 404 Not Found

type NotFoundError struct {
	Message string
}

func NewNotFoundError(msg string) error {
	return &NotFoundError{Message: msg}
}

func (e *NotFoundError) Error() string {
	return e.Message
}

func (NotFoundError) StatusCode() int {
	return http.StatusNotFound
}

func (e *NotFoundError) MarshalJSON() ([]byte, error) {
	return json.Marshal(&ErrorResponse{
		Error: APIError{
			Code:    60802,
			Message: e.Error(),
		},
	})
}

type InternalServerError struct{}

func NewInternalServerError() error {
	return &InternalServerError{}
}

func (*InternalServerError) Error() string {
	return "internal server error"
}

func (InternalServerError) StatusCode() int {
	return http.StatusInternalServerError
}

func (e *InternalServerError) MarshalJSON() ([]byte, error) {
	return json.Marshal(&ErrorResponse{
		Error: APIError{
			Code:    60901,
			Message: e.Error(),
		},
	})
}
