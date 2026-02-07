package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"backend/internal/generated/api/generated"
	"backend/internal/response"
	"backend/internal/service"

	"go.uber.org/zap"
)

// SendOtpEmail looks up the latest OTP from the verification table and sends it via email.
// Called by Better Auth's sendVerificationOTP hook (server-to-server). The OTP never
// travels over the network — it is read directly from the database.
// (POST /auth/otp-email)
func (h *Handlers) SendOtpEmail(w http.ResponseWriter, r *http.Request) {
	var body generated.OtpEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body", "Invalid request body")
		return
	}

	email := string(body.Email)
	if email == "" {
		writeError(w, http.StatusBadRequest, "email_required", "Email address is required")
		return
	}

	// Map the API OTP type to the service OTP type.
	var otpType service.OTPType
	switch body.Type {
	case generated.SignIn:
		otpType = service.OTPTypeSignIn
	case generated.EmailVerification:
		otpType = service.OTPTypeEmailVerification
	case generated.ForgetPassword:
		otpType = service.OTPTypeForgetPassword
	default:
		writeError(w, http.StatusBadRequest, "invalid_type", "Invalid OTP type")
		return
	}

	// Build the identifier that Better Auth uses in the verification table:
	// format: "{type}-otp-{email}"
	identifier := fmt.Sprintf("%s-otp-%s", string(otpType), email)

	otp, err := h.verification.GetRecentOTP(r.Context(), identifier)
	if err != nil {
		h.logger.Error("failed to look up OTP from verification table",
			zap.String("identifier", identifier),
			zap.Error(err),
		)
		// Return 200 to avoid leaking whether the email exists.
		response.WriteJSON(w, http.StatusOK, generated.OtpEmailResponse{})
		return
	}

	if err := h.email.SendOTPEmail(r.Context(), email, otp, otpType); err != nil {
		h.logger.Error("failed to send OTP email",
			zap.String("to", email),
			zap.Error(err),
		)
		response.WriteJSON(w, http.StatusOK, generated.OtpEmailResponse{})
		return
	}

	response.WriteJSON(w, http.StatusOK, generated.OtpEmailResponse{})
}
