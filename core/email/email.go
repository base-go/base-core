package email

import (
	"base/core/config"
	"fmt"
	"sync"
)

var (
	sender Sender
	once   sync.Once
)

type Message struct {
	To      []string
	From    string
	Subject string
	Body    string
	IsHTML  bool
}

type Sender interface {
	Send(msg Message) error
}

// Initialize sets up the email sender based on the configuration
func Initialize(cfg *config.Config) error {
	var err error
	once.Do(func() {
		sender, err = NewSender(cfg)
	})
	return err
}

// Send sends an email using the configured email provider
func Send(msg Message) error {
	if sender == nil {
		return fmt.Errorf("email sender not initialized")
	}
	return sender.Send(msg)
}

// NewEmailSender creates a new email sender based on the configuration
func NewSender(cfg *config.Config) (Sender, error) {
	fmt.Printf("Initializing email sender with provider: %s\n", cfg.EmailProvider)

	switch cfg.EmailProvider {
	case "smtp":
		return NewSMTPSender(cfg)
	case "sendgrid":
		return NewSendGridSender(cfg)
	case "postmark":
		return NewPostmarkSender(cfg)
	case "default":
		return NewDefaultSender(cfg)
	case "":
		fmt.Println("EMAIL_PROVIDER not set, using default sender")
		return NewDefaultSender(cfg)
	default:
		return nil, fmt.Errorf("unsupported email provider: %s", cfg.EmailProvider)
	}
}
