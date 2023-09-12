package mailer

import (
	"fmt"
	"github.com/go-mail/mail/v2"
	"github.com/spf13/viper"
)

const (
	DefaultSender = "support@webgallery.com"
)

type EmailService struct {
	DefaultSender string

	Dialer *mail.Dialer
}

type Email struct {
	From      string
	To        string
	Subject   string
	Plaintext string
	HTML      string
}

type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
}

func DefaultSMTPConfig() SMTPConfig {
	return SMTPConfig{
		Host:     viper.GetString("EMAIL_HOST"),
		Port:     viper.GetInt("EMAIL_PORT"),
		Username: viper.GetString("EMAIL_USERNAME"),
		Password: viper.GetString("EMAIL_PASSWORD"),
	}
}

func NewEmailService(config SMTPConfig) *EmailService {
	es := EmailService{
		Dialer: mail.NewDialer(config.Host, config.Port, config.Username, config.Password),
	}

	return &es
}

func (es *EmailService) SendEmail(email Email) error {
	msg := mail.NewMessage()
	msg.SetHeader("To", email.To)
	es.setFrom(msg, email)
	msg.SetHeader("Subject", email.Subject)

	switch {
	case email.Plaintext != "" && email.HTML != "":
		msg.SetBody("text/plain", email.Plaintext)
		msg.AddAlternative("text/html", email.HTML)
	case email.Plaintext != "":
		msg.SetBody("text/plain", email.Plaintext)
	case email.HTML == "":
		msg.AddAlternative("text/html", email.HTML)
	}

	err := es.Dialer.DialAndSend(msg)
	if err != nil {
		return fmt.Errorf("send: %w", err)
	}
	return nil
}

func (es *EmailService) setFrom(msg *mail.Message, email Email) {
	var from string
	switch {
	case email.From != "":
		from = email.From
	case es.DefaultSender != "":
		from = es.DefaultSender
	default:
		from = DefaultSender
	}

	msg.SetHeader("From", from)

}

func (es *EmailService) ForgotPassword(to, resertURL string) error {
	email := Email{
		Subject:   "Reset your password",
		To:        to,
		Plaintext: "To reset your password, please visit the following link: " + resertURL,
		HTML: `<p> To reset your password, please visit the following link: <a href="` + resertURL + `">` +
			resertURL + `</a></p>`,
	}
	err := es.SendEmail(email)
	if err != nil {
		return fmt.Errorf("forgot password email: %v", err)
	}

	return nil
}
