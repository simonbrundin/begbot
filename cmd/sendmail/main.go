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

	// Provide sample data for the template when sending from CLI
	sampleData := map[string]interface{}{
		"Title":       "TEST: Unik annons — OBS byt data",
		"Price":       "1 234 kr",
		"Valuation":   "10 000 kr",
		"Profit":      "8 766 kr",
		"Discount":    "87%",
		"Description": "DETTA ÄR ETT TESTMEDDELANDE: Byt ut till riktiga värden. Unikt id: 20260219-2",
		"ImageURLs": []string{
			"https://via.placeholder.com/800x600.png?text=Test+Image+1",
			"https://via.placeholder.com/800x600.png?text=Test+Image+2",
		},
		"Link":     "https://example.com/test-annons/2",
		"NewPrice": "12 345 kr",
		"Brand":    "TestBrand",
		"Name":     "TestModel X",
	}

	err = services.SendMailHTMLWithData(emailConfig, cfg.Email.Recipients, subject, "mail.html", sampleData)
	if err != nil {
		panic(err)
	}

	fmt.Println("Mail sent successfully!")
}
