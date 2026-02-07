package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"backend/internal/config"

	"go.uber.org/zap"
)

const plunkSendEndpoint = "https://api.useplunk.com/v1/send"

// safePrefix returns the first n characters of s, or the whole string if shorter
func safePrefix(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

type PlunkSendRequest struct {
	To         string            `json:"to"`
	Subject    string            `json:"subject"`
	Body       string            `json:"body"`
	Subscribed bool              `json:"subscribed"`
	Name       string            `json:"name,omitempty"`
	From       string            `json:"from,omitempty"`
	Reply      string            `json:"reply,omitempty"`
	Headers    map[string]string `json:"headers,omitempty"`
}

type EmailService interface {
	SendInviteEmail(ctx context.Context, to string, inviteURL string, expiresAt time.Time) error
	SendOTPEmail(ctx context.Context, to string, otp string, otpType OTPType) error
}

type emailService struct {
	cfg    config.PlunkConfig
	client *http.Client
	logger *zap.Logger
}

func NewEmailService(cfg config.Config, logger *zap.Logger) EmailService {
	return &emailService{
		cfg:    cfg.Plunk,
		client: &http.Client{Timeout: 30 * time.Second},
		logger: logger,
	}
}

func (s *emailService) SendInviteEmail(ctx context.Context, to string, inviteURL string, expiresAt time.Time) error {
	// Check if Plunk is configured
	if s.cfg.APIKey == "" {
		s.logger.Warn("Plunk API key not configured, skipping email",
			zap.String("to", to),
			zap.String("inviteURL", inviteURL),
		)
		return nil
	}

	data := InviteEmailData{
		Brand:     "BlessThun Food",
		InviteURL: inviteURL,
		ExpiresAt: formatExpiry(expiresAt),
	}

	htmlBody := renderInviteHTML(data)
	textBody := renderInviteText(data)

	payload := PlunkSendRequest{
		To:         to,
		Subject:    "Admin-Einladung",
		Body:       htmlBody,
		Subscribed: false,
		Name:       s.cfg.FromName,
		From:       s.cfg.FromEmail,
		Reply:      s.cfg.ReplyTo,
		Headers: map[string]string{
			"X-Text-Version": textBody,
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal email payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, plunkSendEndpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.cfg.APIKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("plunk returned status %d", resp.StatusCode)
	}

	s.logger.Info("sent invite email",
		zap.String("to", to),
	)

	return nil
}

func (s *emailService) SendOTPEmail(ctx context.Context, to string, otp string, otpType OTPType) error {
	// DEBUG: Log API key info (remove after debugging)
	s.logger.Info("DEBUG: Plunk config",
		zap.Int("api_key_len", len(s.cfg.APIKey)),
		zap.String("api_key_prefix", safePrefix(s.cfg.APIKey, 10)),
	)

	// Check if Plunk is configured
	if s.cfg.APIKey == "" {
		s.logger.Warn("Plunk API key not configured, skipping OTP email",
			zap.String("to", to),
			zap.String("type", string(otpType)),
		)
		return nil
	}

	// OTP expires in 5 minutes (300 seconds) - matches Better Auth config
	const otpExpiresInSeconds = 300

	data := OTPEmailData{
		Brand:       "BlessThun Food",
		Code:        otp,
		CodeTTL:     friendlyTTL(otpExpiresInSeconds),
		SupportNote: "Wir werden dich niemals nach deinem Code fragen.",
	}

	htmlBody := renderOTPHTML(data)
	textBody := renderOTPText(data)
	subject := getOTPSubject(otpType)

	payload := PlunkSendRequest{
		To:         to,
		Subject:    subject,
		Body:       htmlBody,
		Subscribed: false,
		Name:       s.cfg.FromName,
		From:       s.cfg.FromEmail,
		Reply:      s.cfg.ReplyTo,
		Headers: map[string]string{
			"X-Text-Version": textBody,
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal OTP email payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, plunkSendEndpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.cfg.APIKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send OTP email: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// DEBUG: Read response body for error details
		var respBody bytes.Buffer
		respBody.ReadFrom(resp.Body)
		s.logger.Error("DEBUG: Plunk error response",
			zap.Int("status", resp.StatusCode),
			zap.String("body", respBody.String()),
			zap.String("request_body", string(jsonData)),
		)
		return fmt.Errorf("plunk returned status %d", resp.StatusCode)
	}

	s.logger.Info("sent OTP email",
		zap.String("to", to),
		zap.String("type", string(otpType)),
	)

	return nil
}
