# Email Sending

**Rule:**
- Använd `net/smtp` för SMTP-sändning (inga tredjepartslibs)
- Använd StartTLS för säkerhet
- Bygg mail manuellt med MIME headers

**Exception:**
- För high-volume (>1000 mail/dag): Använd SES/SendGrid API

**Exempel:**
```go
func SendEmail(config EmailConfig, to []string, subject, htmlContent string) error {
    addr := config.SMTPHost + ":" + config.SMTPPort
    
    client, err := smtp.Dial(addr)
    if err != nil {
        return err
    }
    defer client.Quit()
    
    err = client.StartTLS(&tls.Config{ServerName: config.SMTPHost})
    // ... Auth, Mail, Rcpt, Data
}
```
