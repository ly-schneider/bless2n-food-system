package middleware

import (
	"backend/internal/errors"
	"backend/internal/http/respond"
	"backend/internal/logger"
	"net/http"
)

// ErrorHandler intercepts *errors.APIError bubbling up the chain.
func ErrorHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// capture panic → 500
		defer func() {
			if rec := recover(); rec != nil {
				logger.L.Errorw("panic", "recover", rec)
				respond.JSON(w, http.StatusInternalServerError,
					errors.FromStatus(http.StatusInternalServerError, "internal error", nil))
			}
		}()

		// run handler
		rw := respond.NewWriter(w) // custom ResponseWriter that stores err
		next.ServeHTTP(rw, r)

		// if handler called rw.SetError(...)
		if apiErr := rw.APIError(); apiErr != nil {
			respond.JSON(w, apiErr.Status, apiErr)
		}
	})
}
