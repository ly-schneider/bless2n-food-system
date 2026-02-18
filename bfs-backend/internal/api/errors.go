package api

import (
	"errors"
	"net/http"

	"backend/internal/generated/api/generated"
	"backend/internal/generated/ent"
	"backend/internal/repository"
	"backend/internal/response"
)

// writeError writes a generated.Error JSON response with the given HTTP status,
// machine-readable code, and human-readable message.
func writeError(w http.ResponseWriter, status int, code, message string) {
	response.WriteJSON(w, status, generated.Error{
		Code:    code,
		Message: message,
	})
}

// writeEntError inspects err and maps it to an appropriate HTTP error response.
//
// Mapping:
//   - ent.IsNotFound / repository.ErrNotFound   -> 404 Not Found
//   - ent.IsConstraintError / repository.ErrConflict -> 409 Conflict
//   - ent.IsValidationError                      -> 400 Bad Request
//   - everything else                            -> 500 Internal Server Error
func writeEntError(w http.ResponseWriter, err error) {
	switch {
	case ent.IsNotFound(err) || errors.Is(err, repository.ErrNotFound):
		writeError(w, http.StatusNotFound, "not_found", "The requested resource was not found.")

	case ent.IsConstraintError(err) || errors.Is(err, repository.ErrConflict):
		writeError(w, http.StatusConflict, "conflict", "The resource already exists or violates a constraint.")

	case ent.IsValidationError(err):
		writeError(w, http.StatusBadRequest, "validation_error", err.Error())

	default:
		writeError(w, http.StatusInternalServerError, "internal_error", "An unexpected error occurred.")
	}
}
