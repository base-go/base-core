package email

import (
	"base/core/config"
	"fmt"

	log "github.com/sirupsen/logrus"
)

type DefaultSender struct{}

func NewDefaultSender(cfg *config.Config) (*DefaultSender, error) {
	return &DefaultSender{}, nil
}

func (s *DefaultSender) Send(msg Message) error {
	log.WithFields(log.Fields{
		"to":      msg.To,
		"from":    msg.From,
		"subject": msg.Subject,
		"isHTML":  msg.IsHTML,
	}).Info("Simulating email send")

	fmt.Println("Email Content:")
	fmt.Println("-------------------")
	fmt.Println(msg.Body)
	fmt.Println("-------------------")

	return nil
}
