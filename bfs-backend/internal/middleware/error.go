package middleware

import (
	"backend/internal/response"
	"context"
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

const (
	ErrorContextKey contextKey = "error_context"
)

type ErrorContext struct {
	RequestID string
	UserID    string
	Path      string
	Method    string
}

func WithErrorContext(ctx context.Context, errorCtx ErrorContext) context.Context {
	return context.WithValue(ctx, ErrorContextKey, errorCtx)
}

func GetErrorContext(ctx context.Context) (ErrorContext, bool) {
	errorCtx, ok := ctx.Value(ErrorContextKey).(ErrorContext)
	return errorCtx, ok
}

type ErrorMiddleware struct {
	logger *zap.Logger
}

func NewErrorMiddleware(logger *zap.Logger) *ErrorMiddleware {
	return &ErrorMiddleware{
		logger: logger,
	}
}

func (e *ErrorMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				e.logger.Error("Handler panicked",
					zap.Any("error", err),
					zap.String("stack", string(debug.Stack())),
					zap.String("method", r.Method),
					zap.String("path", r.URL.Path),
				)

				problem := response.NewProblem(
					http.StatusInternalServerError,
					"Internal Server Error",
					"An unexpected error occurred while processing your request",
				)

				if errorCtx, ok := GetErrorContext(r.Context()); ok {
					problem.Instance = fmt.Sprintf("%s %s", errorCtx.Method, errorCtx.Path)
				}

				response.WriteProblem(w, problem)
			}
		}()

		next.ServeHTTP(w, r)
	}
}

func (e *ErrorMiddleware) HandleError(w http.ResponseWriter, r *http.Request, err error) {
	if err == nil {
		return
	}

	var problem response.ProblemDetails
	errorCtx, hasContext := GetErrorContext(r.Context())

	switch {
	case isValidationError(err):
		problem = e.handleValidationError(err, r)
	case isNotFoundError(err):
		problem = response.NewProblem(
			http.StatusNotFound,
			"Not Found",
			err.Error(),
		)
	case isUnauthorizedError(err):
		problem = response.NewProblem(
			http.StatusUnauthorized,
			"Unauthorized",
			"Authentication is required to access this resource",
		)
	case isForbiddenError(err):
		problem = response.NewProblem(
			http.StatusForbidden,
			"Forbidden",
			"You do not have permission to access this resource",
		)
	case isConflictError(err):
		problem = response.NewProblem(
			http.StatusConflict,
			"Conflict",
			err.Error(),
		)
	case isBadRequestError(err):
		problem = response.NewProblem(
			http.StatusBadRequest,
			"Bad Request",
			err.Error(),
		)
	default:
		// Log unexpected errors with more details
		e.logger.Error("Unexpected error",
			zap.Error(err),
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
		)

		problem = response.NewProblem(
			http.StatusInternalServerError,
			"Internal Server Error",
			"An unexpected error occurred while processing your request",
		)
	}

	// Add instance information if available
	if hasContext {
		problem.Instance = fmt.Sprintf("%s %s", errorCtx.Method, errorCtx.Path)
	}

	response.WriteProblem(w, problem)
}

func (e *ErrorMiddleware) handleValidationError(err error, r *http.Request) response.ProblemDetails {
	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) {
		convertedErrors := response.ConvertValidationErrors(validationErrors)
		instance := ""
		if errorCtx, ok := GetErrorContext(r.Context()); ok {
			instance = fmt.Sprintf("%s %s", errorCtx.Method, errorCtx.Path)
		}
		return response.NewValidationProblem(convertedErrors, instance)
	}

	// Fallback for other validation-like errors
	return response.NewProblem(
		http.StatusBadRequest,
		"Validation Failed",
		err.Error(),
	)
}

// Error type checkers
func isValidationError(err error) bool {
	var validationErrors validator.ValidationErrors
	return errors.As(err, &validationErrors)
}

func isNotFoundError(err error) bool {
	return contains(err.Error(), "not found") || contains(err.Error(), "does not exist")
}

func isUnauthorizedError(err error) bool {
	return contains(err.Error(), "unauthorized") || contains(err.Error(), "unauthenticated")
}

func isForbiddenError(err error) bool {
	return contains(err.Error(), "forbidden") || contains(err.Error(), "access denied")
}

func isConflictError(err error) bool {
	return contains(err.Error(), "already exists") || contains(err.Error(), "conflict")
}

func isBadRequestError(err error) bool {
	return contains(err.Error(), "invalid") || contains(err.Error(), "malformed")
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || (len(s) > len(substr) &&
		(s[:len(substr)+1] == substr+" " || s[len(s)-len(substr)-1:] == " "+substr ||
			indexOf(s, " "+substr+" ") >= 0)))
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

type AppError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Type    string `json:"type"`
}

func (e AppError) Error() string {
	return e.Message
}

func NewNotFoundError(message string) error {
	return AppError{
		Code:    http.StatusNotFound,
		Message: message,
		Type:    "not_found",
	}
}

func NewUnauthorizedError(message string) error {
	return AppError{
		Code:    http.StatusUnauthorized,
		Message: message,
		Type:    "unauthorized",
	}
}

func NewForbiddenError(message string) error {
	return AppError{
		Code:    http.StatusForbidden,
		Message: message,
		Type:    "forbidden",
	}
}

func NewConflictError(message string) error {
	return AppError{
		Code:    http.StatusConflict,
		Message: message,
		Type:    "conflict",
	}
}

func NewBadRequestError(message string) error {
	return AppError{
		Code:    http.StatusBadRequest,
		Message: message,
		Type:    "bad_request",
	}
}
