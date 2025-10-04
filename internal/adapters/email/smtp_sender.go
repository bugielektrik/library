package email

import (
	"bytes"
	"fmt"
	"html/template"
	"net/smtp"

	"go.uber.org/zap"
)

// Config holds SMTP configuration
type Config struct {
	Host     string
	Port     string
	Username string
	Password string
	From     string
}

// SMTPSender sends emails via SMTP
type SMTPSender struct {
	config *Config
	logger *zap.Logger
}

// NewSMTPSender creates a new SMTP email sender
func NewSMTPSender(config *Config, logger *zap.Logger) *SMTPSender {
	return &SMTPSender{
		config: config,
		logger: logger,
	}
}

// Email represents an email message
type Email struct {
	To      []string
	Subject string
	Body    string
	IsHTML  bool
}

// Send sends an email
func (s *SMTPSender) Send(email Email) error {
	auth := smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.Host)

	var contentType string
	if email.IsHTML {
		contentType = "text/html"
	} else {
		contentType = "text/plain"
	}

	msg := fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"Content-Type: %s; charset=UTF-8\r\n"+
		"\r\n"+
		"%s", s.config.From, email.To[0], email.Subject, contentType, email.Body)

	addr := fmt.Sprintf("%s:%s", s.config.Host, s.config.Port)

	if err := smtp.SendMail(addr, auth, s.config.From, email.To, []byte(msg)); err != nil {
		s.logger.Error("failed to send email", zap.Error(err), zap.Strings("to", email.To))
		return fmt.Errorf("failed to send email: %w", err)
	}

	s.logger.Info("email sent successfully", zap.Strings("to", email.To), zap.String("subject", email.Subject))
	return nil
}

// SendTemplate sends an email using a template
func (s *SMTPSender) SendTemplate(to []string, subject, templateName string, data interface{}) error {
	tmpl, err := template.ParseFiles(fmt.Sprintf("internal/adapters/email/templates/%s.html", templateName))
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	var body bytes.Buffer
	if err := tmpl.Execute(&body, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return s.Send(Email{
		To:      to,
		Subject: subject,
		Body:    body.String(),
		IsHTML:  true,
	})
}
