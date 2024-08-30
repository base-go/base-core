package email

import (
	"base/core/config"
	"fmt"
	"net/smtp"
)

type SMTPSender struct {
	host     string
	port     int
	username string
	password string
	from     string
}

func NewSMTPSender(cfg *config.Config) (*SMTPSender, error) {
	return &SMTPSender{
		host:     cfg.SMTPHost,
		port:     cfg.SMTPPort,
		username: cfg.SMTPUsername,
		password: cfg.SMTPPassword,
		from:     cfg.EmailFromAddress,
	}, nil
}

func (s *SMTPSender) Send(msg Message) error {
	auth := smtp.PlainAuth("", s.username, s.password, s.host)
	addr := fmt.Sprintf("%s:%d", s.host, s.port)

	var contentType string
	if msg.IsHTML {
		contentType = "Content-Type: text/html; charset=UTF-8"
	} else {
		contentType = "Content-Type: text/plain; charset=UTF-8"
	}

	message := fmt.Sprintf("To: %s\r\nFrom: %s\r\nSubject: %s\r\n%s\r\n\r\n%s",
		msg.To[0], msg.From, msg.Subject, contentType, msg.Body)

	return smtp.SendMail(addr, auth, s.from, msg.To, []byte(message))
}
