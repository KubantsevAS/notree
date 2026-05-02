package mailer

import (
	"fmt"
	"log"
)

type ConsoleMailer struct{}

func NewConsoleMailer() *ConsoleMailer {
	return &ConsoleMailer{}
}

func (m *ConsoleMailer) SendPasswordReset(email string, token string) error {
	resetLink := fmt.Sprintf("http://localhost:5173/api/auth/reset-password?token=%s", token)

	log.Printf("===========================================")
	log.Printf("Email sent to: %s", email)
	log.Printf("Reset password link: %s", resetLink)
	log.Printf("===========================================")

	return nil
}

func (m *ConsoleMailer) SendVerificationEmail(email string, token string) error {
	resetLink := fmt.Sprintf("http://localhost:5173/api/profile/me/verify-email?token=%s", token)

	log.Printf("===========================================")
	log.Printf("Email sent to: %s", email)
	log.Printf("Verify email link: %s", resetLink)
	log.Printf("===========================================")

	return nil
}
