package respond

import (
	"backend/internal/errors"
	"encoding/json"
	"net/http"
)

type writer struct {
	http.ResponseWriter
	apiErr *errors.APIError
}

func NewWriter(w http.ResponseWriter) *writer { return &writer{ResponseWriter: w} }

func (w *writer) SetError(e *errors.APIError) { w.apiErr = e }
func (w *writer) APIError() *errors.APIError  { return w.apiErr }

// JSON writes data with correct headers.
func JSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
