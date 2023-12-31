package token

import "testing"

func TestGenerateTokenAndSalt(t *testing.T) {
	token, salt, err := GenerateTokenAndSalt(32, 16)
	if err != nil {
		t.Fatal(err)
	}
	// we convert them to strings and it adds more bytes
	if len(token) != 44 {
		t.Errorf("want %d; got %d", 32, len(token))
	}
	if len(salt) != 24 {
		t.Errorf("want %d; got %d", 16, len(salt))
	}
}

func TestHashToken(t *testing.T) {
	token, salt, err := GenerateTokenAndSalt(32, 16)
	if err != nil {
		t.Fatal(err)
	}
	hashedToken := HashToken(token, salt)
	want := 44
	if len(hashedToken) != want {
		t.Errorf("want %d; got %d", want, len(hashedToken))
	}
}

func TestIsValidToken(t *testing.T) {
	token, salt, err := GenerateTokenAndSalt(32, 16)
	if err != nil {
		t.Fatal(err)
	}
	hashedToken := HashToken(token, salt)
	ok := IsValidToken(token, hashedToken, salt)
	if !ok {
		t.Errorf("want %v; got %v", true, ok)
	}
}
