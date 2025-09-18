package response

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

// Response is a generic envelope for successful API responses
type Response[T any] struct {
    Data    T      `json:"data"`
    Message string `json:"message,omitempty"`
}

// Ack is a minimal acknowledgment payload for actions like POST/DELETE
// where returning the full resource is not desired.
type Ack struct {
    Message string `json:"message"`
}

// ProblemDetails represents RFC 9457 Problem Details for HTTP APIs
type ProblemDetails struct {
	Type     string            `json:"type"`
	Title    string            `json:"title"`
	Status   int               `json:"status"`
	Detail   string            `json:"detail,omitempty"`
	Instance string            `json:"instance,omitempty"`
	Errors   []ValidationError `json:"errors,omitempty"`
}

// ValidationError represents field-level validation errors with JSON Pointer references
type ValidationError struct {
	Field   string `json:"field"`   // JSON Pointer (RFC 6901) to the failing field
	Message string `json:"message"` // Human-readable error message
	Value   any    `json:"value,omitempty"`
}

// WriteJSON writes a JSON response with the specified status code
func WriteJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		zap.L().Error("failed to encode response", zap.Error(err))
	}
}

// WriteSuccess writes a successful response using the Response envelope
func WriteSuccess[T any](w http.ResponseWriter, status int, data T, message ...string) {
	response := Response[T]{Data: data}
	if len(message) > 0 {
		response.Message = message[0]
	}
	WriteJSON(w, status, response)
}

// WriteNoContent writes a 204 No Content response with no body per RFC 9110
func WriteNoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// WriteProblem writes an RFC 9457 Problem Details response
func WriteProblem(w http.ResponseWriter, problem ProblemDetails) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(problem.Status)
	if err := json.NewEncoder(w).Encode(problem); err != nil {
		zap.L().Error("failed to encode problem details", zap.Error(err))
	}
}

// WriteError writes an error response using Problem Details
// Deprecated: Use WriteProblem for RFC 9457 compliance
func WriteError(w http.ResponseWriter, status int, message string) {
	problem := ProblemDetails{
		Type:   "about:blank", // Default type per RFC 9457
		Title:  http.StatusText(status),
		Status: status,
		Detail: message,
	}
	WriteProblem(w, problem)
}

// NewProblem creates a new ProblemDetails with defaults
func NewProblem(status int, title, detail string) ProblemDetails {
	return ProblemDetails{
		Type:   "about:blank", // Default type per RFC 9457
		Title:  title,
		Status: status,
		Detail: detail,
	}
}

// NewValidationProblem creates a ProblemDetails for validation errors
func NewValidationProblem(validationErrors []ValidationError, instance string) ProblemDetails {
	return ProblemDetails{
		Type:     "about:blank",
		Title:    "Validation Failed",
		Status:   http.StatusBadRequest,
		Detail:   "The request contains invalid or missing fields",
		Instance: instance,
		Errors:   validationErrors,
	}
}

// ConvertValidationErrors converts go-playground validator errors to our format
func ConvertValidationErrors(validationErrors validator.ValidationErrors) []ValidationError {
	var errors []ValidationError

	for _, err := range validationErrors {
		// Convert struct field path to JSON pointer
		jsonPointer := structFieldToJSONPointer(err.StructNamespace())

		errors = append(errors, ValidationError{
			Field:   jsonPointer,
			Message: getValidationMessage(err),
			Value:   err.Value(),
		})
	}

	return errors
}

// structFieldToJSONPointer converts a struct field path to JSON Pointer format
// Example: "User.Name" -> "#/name", "User.Address.Street" -> "#/address/street"
func structFieldToJSONPointer(structPath string) string {
	if structPath == "" {
		return "#/"
	}

	// Remove the struct name and keep only field path
	parts := strings.Split(structPath, ".")
	if len(parts) <= 1 {
		return "#/"
	}

	// Convert to lowercase and join with /
	fieldParts := parts[1:] // Skip the struct name
	for i, part := range fieldParts {
		fieldParts[i] = strings.ToLower(part)
	}

	return "#/" + strings.Join(fieldParts, "/")
}

// getValidationMessage creates a human-readable message for validation errors
func getValidationMessage(err validator.FieldError) string {
	field := err.Field()

	switch err.Tag() {
	case "required":
		return fmt.Sprintf("Field '%s' is required", field)
	case "email":
		return fmt.Sprintf("Field '%s' must be a valid email address", field)
	case "min":
		return fmt.Sprintf("Field '%s' must be at least %s characters long", field, err.Param())
	case "max":
		return fmt.Sprintf("Field '%s' must be at most %s characters long", field, err.Param())
	case "gte":
		return fmt.Sprintf("Field '%s' must be greater than or equal to %s", field, err.Param())
	case "lte":
		return fmt.Sprintf("Field '%s' must be less than or equal to %s", field, err.Param())
	case "oneof":
		return fmt.Sprintf("Field '%s' must be one of: %s", field, err.Param())
	default:
		return fmt.Sprintf("Field '%s' is invalid", field)
	}
}
