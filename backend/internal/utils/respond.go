package utils

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// ErrorResponse represents an error response structure
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// RespondJSON writes a JSON response with the given status code and data
func RespondJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			http.Error(w, "Failed to encode JSON response", http.StatusInternalServerError)
		}
	}
}

// RespondError writes a JSON error response with the given status code and error
func RespondError(w http.ResponseWriter, statusCode int, err error) {
	response := ErrorResponse{
		Error:   err.Error(),
		Message: http.StatusText(statusCode),
	}
	RespondJSON(w, statusCode, response)
}

// ParseUUID extracts and parses UUID from URL parameter
func ParseUUID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		RespondError(w, http.StatusBadRequest, nil)
		return uuid.Nil, false
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, err)
		return uuid.Nil, false
	}

	return id, true
}
