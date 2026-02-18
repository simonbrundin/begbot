package main

import (
	"fmt"
	"os"

	"begbot/internal/config"
	"begbot/internal/services"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	cfg, err := config.Load("config.yaml")
	if err != nil {
		panic(err)
	}

	fmt.Printf("Config SMTP: host=%s, port=%s, user=%s, from=%s\n",
		cfg.Email.SMTPHost, cfg.Email.SMTPPort, cfg.Email.SMTPUsername, cfg.Email.From)
	fmt.Printf("Recipients: %v\n", cfg.Email.Recipients)

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

	fmt.Println("Mail sent successfully!")
}
