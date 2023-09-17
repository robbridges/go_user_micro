package mailer

import (
	"fmt"
	"github.com/go-mail/mail/v2"
	"github.com/spf13/viper"
	"reflect"
	"strings"
	"testing"
	"the_lonely_road/token"
)

func TestNewEmailService(t *testing.T) {
	mockConfig := SMTPConfig{
		Host:     "testHost",
		Port:     1,
		Username: "rob",
		Password: "secret",
	}

	es := NewEmailService(mockConfig)
	if es.Dialer == nil {
		t.Errorf("Dialer should not be nil")
	}
}

func TestDefaultSMTPConfig(t *testing.T) {
	viper.SetConfigFile("../email.env")
	if err := viper.ReadInConfig(); err != nil {
		t.Fatalf("failed to read config file: %v", err)
	}

	expectedConfig := SMTPConfig{
		Host:     viper.GetString("EMAIL_HOST"),
		Port:     viper.GetInt("EMAIL_PORT"),
		Username: viper.GetString("EMAIL_USERNAME"),
		Password: viper.GetString("EMAIL_PASSWORD"),
	}

	actualConfig := DefaultSMTPConfig()

	if !reflect.DeepEqual(expectedConfig, actualConfig) {
		t.Errorf("Expected config to be %v, but got %v", expectedConfig, actualConfig)
	}
}

func TestSetFrom(t *testing.T) {
	tests := []struct {
		name            string
		defaultSender   string
		emailFrom       string
		expectedResults []string
	}{
		{
			name:            "Email has 'From' field",
			defaultSender:   "default@example.com",
			emailFrom:       "test@example.com",
			expectedResults: []string{"test@example.com"},
		},
		{
			name:            "Email has no 'From' field, default sender is set",
			defaultSender:   "default@example.com",
			emailFrom:       "",
			expectedResults: []string{"default@example.com"},
		},
		{
			name:            "Email has no 'From' field, default sender is not set",
			defaultSender:   "",
			emailFrom:       "",
			expectedResults: []string{DefaultSender},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			es := &EmailService{
				DefaultSender: test.defaultSender,
			}

			email := Email{
				From: test.emailFrom,
			}

			msg := mail.NewMessage()

			es.setFrom(msg, email)

			from := msg.GetHeader("From")

			fromStr := strings.Join(from, ", ")

			matchFound := false
			for _, expectedResult := range test.expectedResults {
				if reflect.DeepEqual(fromStr, expectedResult) {
					matchFound = true
					break
				}
			}

			if !matchFound {
				t.Errorf("Expected From header to be one of %v, but got %s", test.expectedResults, fromStr)
			}
		})
	}
}

func TestSendEmail(t *testing.T) {
	viper.SetConfigFile("../email.env")
	if err := viper.ReadInConfig(); err != nil {
		t.Fatalf("failed to read config file: %v", err)
	}

	cfg := SMTPConfig{
		Host:     viper.GetString("EMAIL_HOST"),
		Port:     viper.GetInt("EMAIL_PORT"),
		Username: viper.GetString("EMAIL_USERNAME"),
		Password: viper.GetString("EMAIL_PASSWORD"),
	}

	emailService := NewEmailService(cfg)

	testEmail := Email{
		To:        "admin@gallery.com",
		Subject:   "Test Email",
		Plaintext: "This is the plaintext content",
		HTML:      "<p>This is the HTML content</p>",
	}

	err := emailService.SendEmail(testEmail)

	if err != nil {
		t.Fatalf("Error sending email: %v", err)
	}
}

func TestEmailService_ForgotPassword(t *testing.T) {
	viper.SetConfigFile("../email.env")
	if err := viper.ReadInConfig(); err != nil {
		t.Fatalf("failed to read config file: %v", err)
	}

	cfg := SMTPConfig{
		Host:     viper.GetString("EMAIL_HOST"),
		Port:     viper.GetInt("EMAIL_PORT"),
		Username: viper.GetString("EMAIL_USERNAME"),
		Password: viper.GetString("EMAIL_PASSWORD"),
	}

	emailService := NewEmailService(cfg)
	passwordToken, _, err := token.GenerateTokenAndSalt(32, 16)
	if err != nil {
		t.Errorf("Error generating token: %v", err)
	}
	url := fmt.Sprintf("localhost:8080/users/password/reset?token=%s", passwordToken)
	err = emailService.ForgotPassword("admin@admin.com", url)
	if err != nil {
		t.Errorf("Error sending email: %v", err)
	}
	if !strings.Contains(url, passwordToken) {
		t.Errorf("Token should be in URL")
	}
}
