package main

import (
	"os"

	"begbot/internal/config"
	"begbot/internal/services"
)

func main() {
	cfg, err := config.Load("config.yaml")
	if err != nil {
		panic(err)
	}

	if len(os.Args) < 2 {
		panic("Subject argument is required")
	}
	subject := os.Args[1]

	emailConfig := services.EmailConfig{
		SMTPHost:     cfg.Email.SMTPHost,
		SMTPPort:     cfg.Email.SMTPPort,
		SMTPUsername: cfg.Email.SMTPUsername,
		SMTPPassword: cfg.Email.SMTPPassword,
		From:         cfg.Email.From,
		Recipients:   cfg.Email.Recipients,
	}

	err = services.SendMailHTML(emailConfig, cfg.Email.Recipients, subject, "mail.html")
	if err != nil {
		panic(err)
	}
}
