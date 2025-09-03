package mocks

import (
	"context"
	"sync"

	"backend/internal/service"
)

// MockEmailService is a mock implementation of EmailService for testing
type MockEmailService struct {
	mu           sync.RWMutex
	sentOTPs     []SentOTP
	sentEmails   []SentEmail
	shouldFail   bool
	failError    error
}

type SentOTP struct {
	Email string
	OTP   string
}

type SentEmail struct {
	To      []string
	Subject string
	Body    string
}

// NewMockEmailService creates a new mock email service
func NewMockEmailService() *MockEmailService {
	return &MockEmailService{
		sentOTPs:   make([]SentOTP, 0),
		sentEmails: make([]SentEmail, 0),
	}
}

// SendOTP implements EmailService interface
func (m *MockEmailService) SendOTP(ctx context.Context, email, otp string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.shouldFail {
		return m.failError
	}

	m.sentOTPs = append(m.sentOTPs, SentOTP{
		Email: email,
		OTP:   otp,
	})

	return nil
}

// SendEmail implements EmailService interface
func (m *MockEmailService) SendEmail(ctx context.Context, req service.SendEmailRequest) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.shouldFail {
		return m.failError
	}

	m.sentEmails = append(m.sentEmails, SentEmail{
		To:      req.To,
		Subject: req.Subject,
		Body:    req.Body,
	})

	return nil
}

// Test helper methods

// GetSentOTPs returns all sent OTPs for testing
func (m *MockEmailService) GetSentOTPs() []SentOTP {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return a copy to avoid race conditions
	result := make([]SentOTP, len(m.sentOTPs))
	copy(result, m.sentOTPs)
	return result
}

// GetSentEmails returns all sent emails for testing
func (m *MockEmailService) GetSentEmails() []SentEmail {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return a copy to avoid race conditions
	result := make([]SentEmail, len(m.sentEmails))
	copy(result, m.sentEmails)
	return result
}

// GetLastSentOTP returns the most recent OTP sent for the given email
func (m *MockEmailService) GetLastSentOTP(email string) *SentOTP {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for i := len(m.sentOTPs) - 1; i >= 0; i-- {
		if m.sentOTPs[i].Email == email {
			return &m.sentOTPs[i]
		}
	}
	return nil
}

// Reset clears all sent emails and OTPs
func (m *MockEmailService) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.sentOTPs = make([]SentOTP, 0)
	m.sentEmails = make([]SentEmail, 0)
	m.shouldFail = false
	m.failError = nil
}

// SetShouldFail configures the mock to fail with the given error
func (m *MockEmailService) SetShouldFail(shouldFail bool, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.shouldFail = shouldFail
	m.failError = err
}