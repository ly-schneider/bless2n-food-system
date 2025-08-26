package apperrors

import (
	"net/http"
)

type APIError struct {
	Status  int    `json:"-"`
	Code    string `json:"code"`    // machine readable
	Message string `json:"message"` // human readable
	Err     error  `json:"-"`       // root cause (optional)
}

func (e *APIError) Error() string { return e.Message }

// Convenience constructors keep handlers readable.
func BadRequest(code, msg string, err error) *APIError {
	return &APIError{Status: http.StatusBadRequest, Code: code, Message: msg, Err: err}
}
func Unauthorized(msg string) *APIError {
	return &APIError{Status: http.StatusUnauthorized, Code: "unauthorized", Message: msg}
}
func Forbidden(msg string) *APIError {
	return &APIError{Status: http.StatusForbidden, Code: "forbidden", Message: msg}
}
func FromStatus(status int, msg string, err error) *APIError {
	return &APIError{Status: status, Code: http.StatusText(status), Message: msg, Err: err}
}
