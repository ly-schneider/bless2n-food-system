package service

import (
	"context"
	"fmt"
	"net/smtp"

	"backend/internal/config"

	"go.uber.org/zap"
)

type EmailService interface {
	SendOTP(ctx context.Context, email, otp string) error
}

type emailService struct {
	config config.SmtpConfig
	logger *zap.Logger
}

func NewEmailService(cfg config.Config, logger *zap.Logger) EmailService {
	return &emailService{
		config: cfg.Smtp,
		logger: logger,
	}
}

func (s *emailService) SendOTP(ctx context.Context, email, otp string) error {
	subject := "Dein Blessthun Code"
	body := fmt.Sprintf(`
Hey,

Dein Anmeldecode für Blessthun ist: %s

Dieser Code läuft in 10 Minuten ab.

Wenn du diesen Code nicht angefordert hast, ignoriere bitte diese E-Mail.

- Dein Blessthun Team
`, otp)

	return s.sendEmail(email, subject, body)
}

func (s *emailService) sendEmail(to, subject, body string) error {
	from := s.config.From
	password := s.config.Password

	// Set up authentication information.
	auth := smtp.PlainAuth("", s.config.Username, password, s.config.Host)

	// Compose the message
	msg := []byte(fmt.Sprintf("To: %s\r\nSubject: %s\r\n\r\n%s\r\n", to, subject, body))

	// Send the email
	addr := fmt.Sprintf("%s:%s", s.config.Host, s.config.Port)
	err := smtp.SendMail(addr, auth, from, []string{to}, msg)
	if err != nil {
		s.logger.Error("failed to send email",
			zap.String("to", to),
			zap.String("subject", subject),
			zap.Error(err))
		return fmt.Errorf("failed to send email: %w", err)
	}

	s.logger.Info("email sent successfully",
		zap.String("to", to),
		zap.String("subject", subject))
	return nil
}
