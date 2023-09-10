package token

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
)

func GenerateTokenAndSalt(tokenLength, saltLength int) (token, salt []byte, err error) {
	// Generate a random token
	token = make([]byte, tokenLength)
	_, err = rand.Read(token)
	if err != nil {
		return nil, nil, err
	}

	// Generate a random salt
	salt = make([]byte, saltLength)
	_, err = rand.Read(salt)
	if err != nil {
		return nil, nil, err
	}

	return token, salt, nil
}

func HashToken(token, salt []byte) []byte {
	// Hash the token with the salt
	hasher := sha256.New()
	hasher.Write(token)
	hasher.Write(salt)
	return hasher.Sum(nil)
}

func IsValidToken(userProvidedToken, storedHashedToken, salt []byte) bool {
	// Hash the user-provided token with the stored salt
	hashedUserToken := HashToken(userProvidedToken, salt)

	// Compare the hashed user-provided token with the stored hashed token
	return bytes.Equal(hashedUserToken, storedHashedToken)
}
