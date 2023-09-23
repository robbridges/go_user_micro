package JWT

import (
	"net/http/httptest"
	"testing"
	"time"
)

func TestGenerateAndValidateJWT(t *testing.T) {
	userID := 123

	// Generate JWT
	token, err := GenerateJWT(userID)
	if err != nil {
		t.Fatalf("generateJWT failed: %v", err)
	}

	if token == "" {
		t.Error("generateJWT produced an empty token")
	}
}
func TestSetAuthCookie(t *testing.T) {
	w := httptest.NewRecorder()
	token := "test-jwt-token"

	SetAuthCookie(w, token)

	// Get the recorded response
	resp := w.Result()

	// Retrieve the "Set-Cookie" header from the response
	cookieHeader := resp.Header.Get("Set-Cookie")
	if cookieHeader == "" {
		t.Error("setAuthCookie did not set the cookie in the response")
	}

}

func TestDeleteAuthCookie(t *testing.T) {

	// Create a response recorder to capture the response
	w := httptest.NewRecorder()

	// Set the cookie
	token := "auth-token"
	SetAuthCookie(w, token)
	cookie := w.Result().Cookies()[0]

	// Check if the cookie is set in the response
	resp := w.Result()
	cookieHeader := resp.Header.Get("Set-Cookie")
	if cookieHeader == "" {
		t.Error("setAuthCookie did not set the cookie in the response")
	}

	DeleteAuthCookie(w, cookie)
	cookie = w.Result().Cookies()[0]
	// cookie expiration should be set to the past
	if cookie.Expires.Before(time.Now()) {
		t.Error("deleteAuthCookie did not set the cookie expiration time to the past")
	}
}
