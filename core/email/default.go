package email

import (
	"base/core/config"
	"fmt"
)

type DefaultSender struct{}

func NewDefaultSender(cfg *config.Config) (*DefaultSender, error) {
	return &DefaultSender{}, nil
}

func (s *DefaultSender) Send(msg Message) error {
	fmt.Printf("Simulating email send - To: %v, From: %s, Subject: %s, IsHTML: %t\n",
		msg.To, msg.From, msg.Subject, msg.IsHTML)

	fmt.Println("Email Content:")
	fmt.Println("-------------------")
	fmt.Println(msg.Body)
	fmt.Println("-------------------")

	return nil
}
