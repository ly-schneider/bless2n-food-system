package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

type JobService struct {
	client *asynq.Client
	logger *zap.Logger
}

func NewJobService(client *asynq.Client, logger *zap.Logger) *JobService {
	return &JobService{
		client: client,
		logger: logger,
	}
}

func (s *JobService) EnqueueEmailVerification(ctx context.Context, payload *EmailVerificationPayload, delay time.Duration) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	task := asynq.NewTask(TypeEmailVerification, data)
	opts := []asynq.Option{
		asynq.ProcessIn(delay),
		asynq.MaxRetry(3),
		asynq.Retention(24 * time.Hour),
	}

	info, err := s.client.EnqueueContext(ctx, task, opts...)
	if err != nil {
		s.logger.Error("failed to enqueue email verification job",
			zap.String("user_id", payload.UserID),
			zap.String("email", payload.Email),
			zap.Error(err))
		return fmt.Errorf("enqueue email verification: %w", err)
	}

	s.logger.Info("email verification job enqueued successfully",
		zap.String("user_id", payload.UserID),
		zap.String("email", payload.Email),
		zap.String("task_id", info.ID),
		zap.Duration("delay", delay))

	return nil
}
