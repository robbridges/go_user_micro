package models

import (
	"reflect"
	"testing"
	"the_lonely_road/token"
	"the_lonely_road/validator"
	"time"
)

func TestUserModelMock_Insert(t *testing.T) {
	mockUser := User{
		ID:        1,
		Email:     "mock@user.com",
		Password:  "mockpassword",
		CreatedAt: time.Now(),
	}

	userModel := UserModelMock{}
	err := userModel.Insert(&mockUser)
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}
	if !reflect.DeepEqual(userModel.DB[0], &mockUser) {
		t.Errorf("Expected user to be inserted")
	}
	// expect error when inserting duplicate user
	err = userModel.Insert(&mockUser)
	if err == nil {
		t.Errorf("Expected error, got nil")
	}

}

func TestUserModelMock_GetByEmail(t *testing.T) {
	mockUser := User{
		ID:        3,
		Email:     "mock@userx.com",
		Password:  "mockpassword",
		CreatedAt: time.Now(),
	}

	userModel := UserModelMock{}
	userModel.DB = append(userModel.DB, &mockUser)

	t.Run("User Found", func(t *testing.T) {
		user, err := userModel.GetByEmail(mockUser.Email)
		if err != nil {
			t.Errorf("Expected no error, got %s", err)
		}
		if !reflect.DeepEqual(user, &mockUser) {
			t.Errorf("Expected user to be returned")
		}
	})
	t.Run("User Not Found", func(t *testing.T) {
		_, err := userModel.GetByEmail("notfound")
		if err == nil && err.Error() != "user not found" {
			t.Errorf("Expected error, got %s", err)
		}
	})
}

func TestUserModelMock_UpdatePassword(t *testing.T) {
	mockUser := User{
		ID:        3,
		Email:     "mock@userx.com",
		Password:  "mockpassword",
		CreatedAt: time.Now(),
	}

	userModel := UserModelMock{}
	userModel.DB = append(userModel.DB, &mockUser)
	t.Run("Happy path", func(t *testing.T) {
		err := userModel.UpdatePassword(int(mockUser.ID), "newpassword")
		if err != nil {
			t.Errorf("Expected no error, got %s", err)
		}
		user, err := userModel.GetByEmail(mockUser.Email)
		if err != nil {
			t.Errorf("Expected no error, got %s", err)
		}
		if user.Password == "newpassword" {
			t.Errorf("Expected password to be updated")
		}
	})
	t.Run("User not found", func(t *testing.T) {
		err := userModel.UpdatePassword(999, "newpassword")
		if err == nil && err.Error() != "user not found" {
			t.Errorf("Expected error, got %s", err)
		}
	})
}

func TestUserModelMock_DeleteUser(t *testing.T) {
	mockUser := User{
		ID:        3,
		Email:     "mock@userx.com",
		Password:  "mockpassword",
		CreatedAt: time.Now(),
	}
	mockUser2 := User{
		ID:        4,
		Email:     "mock@userx3.com",
		Password:  "mockpassword",
		CreatedAt: time.Now(),
	}
	mockModel := UserModelMock{}
	mockModel.DB = append(mockModel.DB, &mockUser, &mockUser2)
	t.Run("Happy path", func(t *testing.T) {
		err := mockModel.DeleteUser(mockUser2.Email)
		if err != nil {
			t.Errorf("Expected no error, got %s", err)
		}
		if len(mockModel.DB) != 1 {
			t.Errorf("Expected mockModel.DB to have length 1, got %d", len(mockModel.DB))
		}
	})
	t.Run("User not found", func(t *testing.T) {
		err := mockModel.DeleteUser("notfound")
		if err == nil && err.Error() != "user not found" {
			t.Errorf("Expected error, got %s", err)
		}
	})
}

func TestUserModelMock_Authenticate(t *testing.T) {
	mockUser := User{
		ID:        3,
		Email:     "mock@userx.com",
		Password:  "mockpassword",
		CreatedAt: time.Now(),
	}
	// need to get a copy of the plaintext before the hash is created
	passwordBefore := mockUser.Password
	userModel := UserModelMock{}
	err := userModel.Insert(&mockUser)
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}
	// compare user hashed password to it's plaintext
	t.Run("Happy path", func(t *testing.T) {
		user, err := userModel.Authenticate(mockUser.Email, passwordBefore)
		if err != nil {
			t.Errorf("Expected no error, got %s", err)
		}
		if user == nil {
			t.Errorf("Expected user to be returned")
		}
		if user.Password == passwordBefore {
			t.Errorf("Expected password to be encrypted")
		}
	})
	t.Run("Wrong password", func(t *testing.T) {
		_, err := userModel.Authenticate(mockUser.Email, "wrongpassword")
		if err == nil && err.Error() != "compare() error: crypto/bcrypt: hashedPassword is not the hash of the given password" {
			t.Errorf("Expected error, got %s", err)
		}
	})
	t.Run("User not found", func(t *testing.T) {
		_, err := userModel.Authenticate("notfound", "notfound")
		if err == nil && err.Error() != "user not found" {
			t.Errorf("Expected error, got %s", err)
		}
	})
}

func TestUserModelMock_EnterPasswordHash(t *testing.T) {
	mockUser := User{
		ID:        3,
		Email:     "hashme@example.com",
		Password:  "mockpassword",
		CreatedAt: time.Now(),
	}
	userModel := UserModelMock{}
	err := userModel.Insert(&mockUser)
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}
	t.Run("Happy path", func(t *testing.T) {
		if !mockUser.PasswordResetExpiry.IsZero() {
			t.Errorf("Expected PasswordResetExpiry to  be unset, got %s", mockUser.PasswordResetExpiry)
		}
		if mockUser.PasswordResetHashToken != "" || mockUser.PasswordResetSalt != "" {
			t.Errorf("Expected PasswordResetToken and PasswordResetSalt to be unset, got %s and %s", mockUser.PasswordResetHashToken, mockUser.PasswordResetSalt)
		}
		passwordHash, salt, err := token.GenerateTokenAndSalt(32, 16)
		if err != nil {
			t.Fatal(err)
		}
		token.HashToken(passwordHash, salt)
		err = userModel.EnterPasswordHash(mockUser.Email, passwordHash, salt)
		if err != nil {
			t.Errorf("Expected no error, got %s", err)
		}
		if mockUser.PasswordResetExpiry.IsZero() {
			t.Errorf("Expected PasswordResetExpiry to be set, got %s", mockUser.PasswordResetExpiry)
		}
		if mockUser.PasswordResetHashToken == "" || mockUser.PasswordResetSalt == "" {
			t.Errorf("Expected PasswordResetToken and PasswordResetSalt to be set, got %s and %s", mockUser.PasswordResetHashToken, mockUser.PasswordResetSalt)
		}
	})

	t.Run("User not found", func(t *testing.T) {
		err = userModel.EnterPasswordHash("notfound", "notfound", "notfound")
		if err == nil && err.Error() != "user not found" {
			t.Errorf("Expected error, got %s", err)
		}
	})
}
func TestUserModelMock_ConsumePasswordReset(t *testing.T) {
	mockUser := User{
		ID:        3,
		Email:     "test@example.com",
		Password:  "mockpassword",
		CreatedAt: time.Now(),
	}
	userModel := UserModelMock{}
	err := userModel.Insert(&mockUser)
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}
	t.Run("Happy path", func(t *testing.T) {
		passwordHash, salt, err := token.GenerateTokenAndSalt(32, 16)
		if err != nil {
			t.Fatal(err)
		}
		token.HashToken(passwordHash, salt)
		err = userModel.EnterPasswordHash(mockUser.Email, passwordHash, salt)
		if err != nil {
			t.Errorf("Expected no error, got %s", err)
		}
		if mockUser.PasswordResetExpiry.IsZero() {
			t.Errorf("Expected PasswordResetExpiry to be set, got %s", mockUser.PasswordResetExpiry)
		}
		if mockUser.PasswordResetHashToken == "" || mockUser.PasswordResetSalt == "" {
			t.Errorf("Expected PasswordResetToken and PasswordResetSalt to be set, got %s and %s", mockUser.PasswordResetHashToken, mockUser.PasswordResetSalt)
		}
		err = userModel.ConsumePasswordReset(mockUser.Email)
		if err != nil {
			t.Errorf("Expected no error, got %s", err)
		}
		if !mockUser.PasswordResetExpiry.IsZero() {
			t.Errorf("Expected PasswordResetExpiry to be unset, got %s", mockUser.PasswordResetExpiry)
		}
		if mockUser.PasswordResetHashToken != "" || mockUser.PasswordResetSalt != "" {
			t.Errorf("Expected PasswordResetToken and PasswordResetSalt to be unset, got %s and %s", mockUser.PasswordResetHashToken, mockUser.PasswordResetSalt)
		}
	})
	t.Run("User not found", func(t *testing.T) {
		err = userModel.ConsumePasswordReset("notfound")
		if err == nil && err.Error() != "user not found" {
			t.Errorf("Expected error, got %s", err)
		}
	})
}

func TestValidateEmail(t *testing.T) {
	t.Run("Happy path", func(t *testing.T) {
		v := validator.New()
		ValidateEmail(v, "AverygoodEmail@great.com")
		if !v.Valid() {
			t.Errorf("Expected validator to be valid, got invalid")
		}
	})
	t.Run("Email too short", func(t *testing.T) {
		v := validator.New()
		ValidateEmail(v, "x@c")
		if v.Valid() {
			t.Errorf("Expected validator to be invalid, got valid")
		}
	})
}

func TestValidatePasswordPlaintext(t *testing.T) {
	t.Run("Happy path", func(t *testing.T) {
		v := validator.New()
		ValidatePasswordPlaintext(v, "averygoodpassword")
		if !v.Valid() {
			t.Errorf("Expected validator to be valid, got invalid")
		}
	})
	t.Run("Password too short", func(t *testing.T) {
		v := validator.New()
		ValidatePasswordPlaintext(v, "abc")
		if v.Valid() {
			t.Errorf("Expected validator to be invalid, got valid")
		}
	})
}

func TestValidateUser(t *testing.T) {
	t.Run("Happy path", func(t *testing.T) {
		v := validator.New()
		user := User{
			Email:    "AverygoodEmail",
			Password: "averygoodpassword",
		}
		ValidateUser(v, &user)
		if !v.Valid() {
			t.Errorf("Expected validator to be valid, got invalid")
		}
	})
	t.Run("Email too short", func(t *testing.T) {
		v := validator.New()
		user := User{
			Email:    "bad",
			Password: "averygoodpassword",
		}
		ValidateUser(v, &user)
		if v.Valid() {
			t.Errorf("Expected validator to not be valid")
		}
	})
	t.Run("Password too short", func(t *testing.T) {
		v := validator.New()
		user := User{
			Email:    "averygoodemail",
			Password: "",
		}
		ValidateUser(v, &user)
		if v.Valid() {
			t.Errorf("Expected validator to not be valid")
		}
	})
}
