package mailer

import (
	"context"
	"fmt"
	"log"
)

type Mailer interface {
	SendPasswordReset(ctx context.Context, email string, token string) error
	SendVerificationEmail(ctx context.Context, email string, token string) error
}

type ConsoleMailer struct{}

func NewConsoleMailer() *ConsoleMailer {
	return &ConsoleMailer{}
}

func (m *ConsoleMailer) SendPasswordReset(ctx context.Context, email string, token string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		resetLink := fmt.Sprintf("http://localhost:5173/api/auth/reset-password?token=%s", token)

		log.Printf("===========================================")
		log.Printf("Email sent to: %s", email)
		log.Printf("Reset password link: %s", resetLink)
		log.Printf("===========================================")

		return nil
	}
}

func (m *ConsoleMailer) SendVerificationEmail(ctx context.Context, email string, token string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		resetLink := fmt.Sprintf("http://localhost:5173/api/profile/me/verify-email?token=%s", token)

		log.Printf("===========================================")
		log.Printf("Email sent to: %s", email)
		log.Printf("Verify email link: %s", resetLink)
		log.Printf("===========================================")

		return nil
	}
}
