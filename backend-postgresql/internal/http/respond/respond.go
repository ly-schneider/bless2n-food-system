package respond

import (
	"backend/internal/apperrors"
	"encoding/json"
	"net/http"
)

type writer struct {
	http.ResponseWriter
	apiErr *apperrors.APIError
}

func NewWriter(w http.ResponseWriter) *writer { return &writer{ResponseWriter: w} }

func (w *writer) WriteError(e *apperrors.APIError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(e.Status)
	_ = json.NewEncoder(w).Encode(e)
}

func (w *writer) APIError() *apperrors.APIError { return w.apiErr }

func JSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if v != nil {
		_ = json.NewEncoder(w).Encode(v)
	}
}
