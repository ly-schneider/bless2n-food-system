package service

import (
	"backend/internal/config"
	"context"
	"fmt"

	"github.com/mailgun/mailgun-go/v5"
)

type EmailService interface {
	SendVerificationEmail(ctx context.Context, to, name string, code string) error
}

type emailService struct {
	mg        *mailgun.Client
	domain    string
	fromEmail string
	fromName  string
}

func NewEmailService(cfg config.MailgunConfig) EmailService {
	mg := mailgun.NewMailgun(cfg.APIKey)
	return &emailService{
		mg:        mg,
		domain:    cfg.Domain,
		fromEmail: cfg.FromEmail,
		fromName:  cfg.FromName,
	}
}

func (s *emailService) SendVerificationEmail(ctx context.Context, to, name string, code string) error {
	subject := "Verify Your Email Address"

	htmlBody := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Email Verification</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h1 style="color: #2c3e50;">Verify Your Email Address</h1>
        <p>Hi %s,</p>
        <p>Thank you for signing up! Please use the following verification code to complete your registration:</p>
        <div style="background-color: #f8f9fa; padding: 20px; text-align: center; margin: 20px 0; border-radius: 5px;">
            <h2 style="color: #2c3e50; margin: 0; font-size: 36px; letter-spacing: 5px;">%s</h2>
        </div>
        <p>This verification code will expire in 15 minutes.</p>
        <p>If you didn't create an account with us, please ignore this email.</p>
        <p>Best regards,<br>The Rentro Team</p>
    </div>
</body>
</html>`, name, code)

	textBody := fmt.Sprintf(`
Hi %s,

Thank you for signing up! Please use the following verification code to complete your registration:

%s

This verification code will expire in 15 minutes.

If you didn't create an account with us, please ignore this email.

Best regards,
The Rentro Team`, name, code)

	message := mailgun.NewMessage(
		s.domain,
		fmt.Sprintf("%s <%s>", s.fromName, s.fromEmail),
		subject,
		textBody,
		to,
	)
	message.SetHTML(htmlBody)

	_, err := s.mg.Send(ctx, message)
	if err != nil {
		return fmt.Errorf("failed to send verification email: %w", err)
	}

	return nil
}
