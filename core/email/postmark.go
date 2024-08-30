package email

import (
	"base/core/config"

	"github.com/keighl/postmark"
)

type PostmarkSender struct {
	client *postmark.Client
	from   string
}

func NewPostmarkSender(cfg *config.Config) (*PostmarkSender, error) {
	client := postmark.NewClient(cfg.PostmarkServerToken, cfg.PostmarkAccountToken)
	return &PostmarkSender{
		client: client,
		from:   cfg.EmailFromAddress,
	}, nil
}

func (s *PostmarkSender) Send(msg Message) error {
	email := postmark.Email{
		From:     s.from,
		To:       msg.To[0],
		Subject:  msg.Subject,
		TextBody: msg.Body,
		HtmlBody: msg.Body,
	}

	if !msg.IsHTML {
		email.HtmlBody = ""
	} else {
		email.TextBody = ""
	}

	_, err := s.client.SendEmail(email)
	return err
}
