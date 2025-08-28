package domain

type TokenType string

const (
	TokenTypeLogin         TokenType = "login"
	TokenTypePasswordReset TokenType = "password_reset"
)
