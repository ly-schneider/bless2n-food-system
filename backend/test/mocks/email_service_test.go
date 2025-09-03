package mocks

import (
	"context"
	"errors"
	"testing"

	"backend/internal/service"

	"github.com/stretchr/testify/assert"
)

func TestMockEmailService_SendOTP(t *testing.T) {
	mock := NewMockEmailService()

	// Test successful OTP sending
	err := mock.SendOTP(context.Background(), "test@example.com", "123456")
	assert.NoError(t, err)

	// Verify OTP was recorded
	sentOTPs := mock.GetSentOTPs()
	assert.Len(t, sentOTPs, 1)
	assert.Equal(t, "test@example.com", sentOTPs[0].Email)
	assert.Equal(t, "123456", sentOTPs[0].OTP)

	// Test GetLastSentOTP
	lastOTP := mock.GetLastSentOTP("test@example.com")
	assert.NotNil(t, lastOTP)
	assert.Equal(t, "123456", lastOTP.OTP)

	// Test failure mode
	expectedErr := errors.New("email service down")
	mock.SetShouldFail(true, expectedErr)
	
	err = mock.SendOTP(context.Background(), "test2@example.com", "654321")
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
}

func TestMockEmailService_SendEmail(t *testing.T) {
	mock := NewMockEmailService()

	req := service.SendEmailRequest{
		To:      []string{"user1@example.com", "user2@example.com"},
		Subject: "Test Subject",
		Body:    "Test Body",
	}

	err := mock.SendEmail(context.Background(), req)
	assert.NoError(t, err)

	// Verify email was recorded
	sentEmails := mock.GetSentEmails()
	assert.Len(t, sentEmails, 1)
	assert.Equal(t, req.To, sentEmails[0].To)
	assert.Equal(t, req.Subject, sentEmails[0].Subject)
	assert.Equal(t, req.Body, sentEmails[0].Body)
}

func TestMockEmailService_Reset(t *testing.T) {
	mock := NewMockEmailService()

	// Send some test data
	mock.SendOTP(context.Background(), "test@example.com", "123456")
	mock.SendEmail(context.Background(), service.SendEmailRequest{
		To:      []string{"user@example.com"},
		Subject: "Test",
		Body:    "Test Body",
	})

	// Verify data exists
	assert.Len(t, mock.GetSentOTPs(), 1)
	assert.Len(t, mock.GetSentEmails(), 1)

	// Reset and verify data is cleared
	mock.Reset()
	assert.Len(t, mock.GetSentOTPs(), 0)
	assert.Len(t, mock.GetSentEmails(), 0)
}