package service

type EmailService interface {
}

type emailService struct {
}

func NewEmailService() EmailService {
	return &emailService{}
}
