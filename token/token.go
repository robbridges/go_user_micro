package token

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
)

func GenerateTokenAndSalt(tokenLength, saltLength int) (token, salt string, err error) {
	// Generate a random token
	tokenBytes := make([]byte, tokenLength)
	_, err = rand.Read(tokenBytes)
	if err != nil {
		return "", "", err
	}
	token = base64.URLEncoding.EncodeToString(tokenBytes)

	// Generate a random salt
	saltBytes := make([]byte, saltLength)
	_, err = rand.Read(saltBytes)
	if err != nil {
		return "", "", err
	}
	salt = base64.URLEncoding.EncodeToString(saltBytes)

	return token, salt, nil
}

func HashToken(token, salt string) string {
	// Decode Base64-encoded token and salt back to bytes
	tokenBytes, err := base64.URLEncoding.DecodeString(token)
	if err != nil {
		// Handle decoding error
		return ""
	}

	saltBytes, err := base64.URLEncoding.DecodeString(salt)
	if err != nil {
		// Handle decoding error
		return ""
	}

	// Hash the token with the salt
	hasher := sha256.New()
	hasher.Write(tokenBytes)
	hasher.Write(saltBytes)

	// Encode the hash as a Base64 string
	hash := base64.URLEncoding.EncodeToString(hasher.Sum(nil))

	return hash
}
func IsValidToken(userProvidedToken, storedHashedToken, salt string) bool {
	// Decode Base64-encoded user-provided token and salt back to bytes
	userTokenBytes, err := base64.URLEncoding.DecodeString(userProvidedToken)
	if err != nil {
		// Handle decoding error
		return false
	}

	saltBytes, err := base64.URLEncoding.DecodeString(salt)
	if err != nil {
		// Handle decoding error
		return false
	}

	// Hash the user-provided token with the stored salt
	hasher := sha256.New()
	hasher.Write(userTokenBytes)
	hasher.Write(saltBytes)
	hashedUserToken := base64.URLEncoding.EncodeToString(hasher.Sum(nil))

	// Compare the hashed user-provided token with the stored hashed token
	return hashedUserToken == storedHashedToken
}
