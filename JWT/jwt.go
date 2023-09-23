package JWT

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"time"
)

func GenerateJWT(userID int) (string, error) {
	// Prepare the JWT payload
	payload := fmt.Sprintf(`{"user_id":%d,"exp":%d}`, userID, time.Now().Add(time.Hour*24).Unix())

	// Create a signature using HMAC-SHA256
	secretKey := []byte("your-secret-key")
	hmac256 := hmac.New(sha256.New, secretKey)
	hmac256.Write([]byte(payload))
	signature := hmac256.Sum(nil)

	// Encode the payload and signature as base64
	encodedPayload := base64.StdEncoding.EncodeToString([]byte(payload))
	encodedSignature := base64.StdEncoding.EncodeToString(signature)

	// Combine the encoded payload and signature with a dot separator
	token := fmt.Sprintf("%s.%s", encodedPayload, encodedSignature)
	return token, nil
}

func SetAuthCookie(w http.ResponseWriter, token string) {
	cookie := &http.Cookie{
		Name:     "auth_token",
		Value:    token,
		Expires:  time.Now().Add(time.Hour * 24), // Same as token expiration time
		HttpOnly: true,
		Secure:   true, // Set to true if using HTTPS
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
	}

	http.SetCookie(w, cookie)
}

func DeleteAuthCookie(w http.ResponseWriter, cookie *http.Cookie) {
	cookie = &http.Cookie{
		Name:     "auth_token",
		Value:    "",
		Expires:  time.Now().Add(time.Hour * -24).UTC(), // Same as token expiration time
		HttpOnly: true,
		Secure:   true, // Set to true if using HTTPS
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
	}
	cookie.MaxAge = -1
	http.SetCookie(w, cookie)
}
