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
}

func NewEmailService(cfg config.Config) EmailService {
	return &emailService{
		config: cfg.Smtp,
	}
}

func (s *emailService) SendOTP(ctx context.Context, email, otp string) error {
	subject := "Dein BlessThun Code"
	body := fmt.Sprintf(`
Hey,

Dein Anmeldecode für BlessThun ist: %s

Dieser Code läuft in 10 Minuten ab.

Wenn du diesen Code nicht angefordert hast, ignoriere bitte diese E-Mail.

- Dein BlessThun Team
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
		zap.L().Error("failed to send email",
			zap.String("to", to),
			zap.String("subject", subject),
			zap.Error(err))
		return fmt.Errorf("failed to send email: %w", err)
	}

	zap.L().Info("email sent successfully",
		zap.String("to", to),
		zap.String("subject", subject))
	return nil
}
