package JWT

import (
	"net/http/httptest"
	"testing"
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
