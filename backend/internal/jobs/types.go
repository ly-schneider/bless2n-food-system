package jobs

const (
	TypeEmailVerification = "email:verification"
)

type EmailVerificationPayload struct {
	UserID      string `json:"user_id"`
	Email       string `json:"email"`
	FirstName   string `json:"first_name"`
	Token       string `json:"token"`
	RequestedAt int64  `json:"requested_at"`
}
