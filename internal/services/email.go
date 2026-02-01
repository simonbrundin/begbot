package services

import (
	"crypto/tls"
	"net/smtp"
	"os"
	"path/filepath"
	"strings"
)

type EmailConfig struct {
	SMTPHost     string
	SMTPPort     string
	SMTPUsername string
	SMTPPassword string
	From         string
	Recipients   []string
}

func LoadEmailHTML(fileName string) (string, error) {
	filePath := filepath.Join(".", fileName)
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func SendEmail(config EmailConfig, to []string, subject, htmlContent string) error {
	auth := smtp.PlainAuth("", config.SMTPUsername, config.SMTPPassword, config.SMTPHost)

	msg := "To: " + strings.Join(to, ",") + "\r\n"
	msg += "From: " + config.From + "\r\n"
	msg += "Subject: " + subject + "\r\n"
	msg += "MIME-Version: 1.0\r\n"
	msg += "Content-Type: text/html; charset=UTF-8\r\n"
	msg += "\r\n"
	msg += htmlContent

	addr := config.SMTPHost + ":" + config.SMTPPort

	client, err := smtp.Dial(addr)
	if err != nil {
		return err
	}

	err = client.StartTLS(&tls.Config{ServerName: config.SMTPHost})
	if err != nil {
		return err
	}

	err = client.Auth(auth)
	if err != nil {
		return err
	}

	err = client.Mail(config.From)
	if err != nil {
		return err
	}

	for _, t := range to {
		err = client.Rcpt(t)
		if err != nil {
			return err
		}
	}

	w, err := client.Data()
	if err != nil {
		return err
	}

	_, err = w.Write([]byte(msg))
	if err != nil {
		return err
	}

	err = w.Close()
	if err != nil {
		return err
	}

	err = client.Quit()
	if err != nil {
		return err
	}

	return nil
}

func SendMailHTML(config EmailConfig, to []string, subject, htmlFileName string) error {
	htmlContent, err := LoadEmailHTML(htmlFileName)
	if err != nil {
		return err
	}

	return SendEmail(config, to, subject, htmlContent)
}
