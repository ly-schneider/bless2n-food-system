package jobs

import (
	"backend/internal/model"
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

// EmailSender defines the interface for sending verification emails
type EmailSender interface {
	SendVerificationCode(ctx context.Context, userID model.NanoID14, email, name string) error
}

type JobHandlers struct {
	logger      *zap.Logger
	emailSender EmailSender
}

func NewJobHandlers(logger *zap.Logger, emailSender EmailSender) *JobHandlers {
	return &JobHandlers{
		logger:      logger,
		emailSender: emailSender,
	}
}

func (h *JobHandlers) HandleEmailVerification(ctx context.Context, t *asynq.Task) error {
	var payload EmailVerificationPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		h.logger.Error("failed to unmarshal email verification payload",
			zap.Error(err),
			zap.String("task_id", t.ResultWriter().TaskID()))
		return fmt.Errorf("unmarshal payload: %w", err)
	}

	h.logger.Info("processing email verification job",
		zap.String("user_id", payload.UserID),
		zap.String("email", payload.Email),
		zap.String("task_id", t.ResultWriter().TaskID()))

	if err := h.sendVerificationEmail(payload); err != nil {
		h.logger.Error("failed to send verification email",
			zap.String("user_id", payload.UserID),
			zap.String("email", payload.Email),
			zap.Error(err))
		return fmt.Errorf("send verification email: %w", err)
	}

	h.logger.Info("email verification job completed successfully",
		zap.String("user_id", payload.UserID),
		zap.String("email", payload.Email))

	return nil
}

func (h *JobHandlers) sendVerificationEmail(payload EmailVerificationPayload) error {
	// Convert string UserID to NanoID14
	userID := model.NanoID14(payload.UserID)

	// Use the email sender to send the actual email with a 6-digit code
	err := h.emailSender.SendVerificationCode(context.Background(), userID, payload.Email, payload.FirstName)
	if err != nil {
		h.logger.Error("failed to send verification code",
			zap.String("user_id", payload.UserID),
			zap.String("email", payload.Email),
			zap.Error(err))
		return fmt.Errorf("failed to send verification code: %w", err)
	}

	h.logger.Info("verification email sent successfully",
		zap.String("to", payload.Email),
		zap.String("user_id", payload.UserID))

	return nil
}
