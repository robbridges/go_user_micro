package mailer

import (
	"github.com/go-mail/mail/v2"
	"reflect"
	"strings"
	"testing"
)

func TestNewEmailService(t *testing.T) {
	mockConfig := SMTPConfig{
		Host:     "testhost",
		Port:     1,
		Username: "rob",
		Password: "secret",
	}

	es := NewEmailService(mockConfig)
	if es.Dialer == nil {
		t.Errorf("Dialer should not be nil")
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
