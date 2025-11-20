package email

import (
	"context"
	"fmt"

	"base/core/errors"
	"base/core/logger"
)

// EmailProviderType represents the email provider type (avoiding name conflict)
type EmailProviderType string

const (
	EmailProviderSMTP     EmailProviderType = "smtp"
	EmailProviderSendGrid EmailProviderType = "sendgrid"
	EmailProviderPostmark EmailProviderType = "postmark"
)

// SimpleManager wraps the existing email system with better patterns
type SimpleManager struct {
	logger logger.Logger
}

// NewSimpleManager creates a new simple email manager
func NewSimpleManager(log logger.Logger) *SimpleManager {
	return &SimpleManager{
		logger: log,
	}
}

// Send sends an email using the existing global sender
func (m *SimpleManager) Send(ctx context.Context, message Message) error {
	m.logger.Info("Sending email",
		logger.String("to", fmt.Sprintf("%v", message.To)),
		logger.String("subject", message.Subject),
	)

	// Use the existing global Send function
	err := Send(message)
	if err != nil {
		m.logger.Error("Failed to send email",
			logger.String("error", err.Error()),
		)
		return errors.Wrap(err, errors.CodeEmailSend, "failed to send email")
	}

	m.logger.Info("Email sent successfully")
	return nil
}

// SendWithContext is an alias for Send with context (for future expansion)
func (m *SimpleManager) SendWithContext(ctx context.Context, message Message) error {
	return m.Send(ctx, message)
}
