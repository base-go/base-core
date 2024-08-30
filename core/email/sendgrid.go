package email

import (
	"base/core/config"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type SendGridSender struct {
	client *sendgrid.Client
	from   string
}

func NewSendGridSender(cfg *config.Config) (*SendGridSender, error) {
	client := sendgrid.NewSendClient(cfg.SendGridAPIKey)
	return &SendGridSender{
		client: client,
		from:   cfg.EmailFromAddress,
	}, nil
}

func (s *SendGridSender) Send(msg Message) error {
	from := mail.NewEmail("", s.from)
	to := mail.NewEmail("", msg.To[0])
	content := mail.NewContent("text/plain", msg.Body)
	if msg.IsHTML {
		content = mail.NewContent("text/html", msg.Body)
	}

	email := mail.NewV3MailInit(from, msg.Subject, to, content)

	_, err := s.client.Send(email)
	return err
}
