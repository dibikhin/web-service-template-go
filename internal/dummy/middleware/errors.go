package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
)

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
	return json.Marshal(&struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}{
		Code:    60901,
		Message: e.Error(),
	})
}

type ValidationError struct {
	Message string
}

func NewValidationError(msg string) error {
	return &ValidationError{Message: msg}
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error: %s", e.Message)
}

func (ValidationError) StatusCode() int {
	return http.StatusBadRequest
}

func (e *ValidationError) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}{
		Code:    60801,
		Message: e.Error(),
	})
}
