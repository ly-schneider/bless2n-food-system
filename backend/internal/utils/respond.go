package utils

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
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
		Message: http.StatusText(statusCode),
	}

	if err != nil {
		response.Error = err.Error()
	}

	RespondJSON(w, statusCode, response)
}

// ParseUUID extracts and parses NanoID from URL parameter
func ParseUUID(w http.ResponseWriter, r *http.Request) (string, bool) {
	id := chi.URLParam(r, "id")
	if id == "" {
		RespondError(w, http.StatusBadRequest, nil)
		return "", false
	}

	// For NanoID we don't need to parse, just validate
	if err := Validate("id", id); err != nil {
		RespondError(w, http.StatusBadRequest, err)
		return "", false
	}

	return id, true
}
