package models

import (
	"reflect"
	"testing"
	"the_lonely_road/data"
	"the_lonely_road/token"
	"time"
)

type App struct {
	userModel *UserModel
}

func TestUserModel_Insert(t *testing.T) {
	mockUser := User{
		ID:                     1,
		Email:                  "justatest@test.com",
		Password:               "mockpassword",
		CreatedAt:              time.Now(),
		PasswordResetSalt:      "test_salt",
		PasswordResetExpiry:    time.Now().Add(24 * time.Hour), // Set an expiration time
		PasswordResetHashToken: "test_token",
	}

	cfg := data.TestPostgresConfig()
	db, err := data.Open(cfg)
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}
	defer db.Close()

	t.Run("Insert User happy path", func(t *testing.T) {

		userModel := &UserModel{DB: db}
		err = userModel.Insert(&mockUser)
		if err != nil {
			t.Errorf("Expected no error, got %s", err)
		}
		foundUser, err := userModel.GetByEmail(mockUser.Email)
		if err != nil {
			t.Errorf("Expected no error, got %s", err)
		}
		if reflect.DeepEqual(foundUser, &mockUser) {
			t.Errorf("retrieved user does not match inserted user")
		}

		err = userModel.DeleteUser(mockUser.Email)
		if err != nil {
			t.Errorf("Expected no error, got %s", err)
		}
	})

	t.Run("Insert duplicate user", func(t *testing.T) {

		userModel := &UserModel{DB: db}
		err = userModel.Insert(&mockUser)
		if err != nil {
			t.Errorf("Expected no error, got %s", err)
		}
		err = userModel.Insert(&mockUser)
		if err == nil {
			t.Errorf("Expected error, got nil")
		}
		err = userModel.DeleteUser(mockUser.Email)
	})

}

func TestUserModel_GetByEmail(t *testing.T) {

	cfg := data.TestPostgresConfig()
	db, err := data.Open(cfg)
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}
	defer db.Close()

	userModel := &UserModel{DB: db}
	t.Run("User Found", func(t *testing.T) {
		userToInsert := User{
			ID:                     1,
			Email:                  "testuser@localhost",
			Password:               "mockpassword",
			CreatedAt:              time.Now(),
			PasswordResetSalt:      "test_salt",
			PasswordResetExpiry:    time.Now().Add(24 * time.Hour), // Set an expiration time
			PasswordResetHashToken: "test_token",
		}
		err := userModel.Insert(&userToInsert)
		if err != nil {
			t.Errorf("Expected no error, got %s", err)
		}

		user, err := userModel.GetByEmail(userToInsert.Email)
		if err != nil {
			t.Errorf("Expected no error, got %s", err)
		}

		if reflect.DeepEqual(user, &userToInsert) {
			t.Errorf("Expected user to be returned")
		}
		err = userModel.DeleteUser(userToInsert.Email)
	})
	t.Run("User Not Found", func(t *testing.T) {
		_, err := userModel.GetByEmail("notfound")
		if err == nil && err.Error() != "record not found" {
			t.Errorf("Expected error, got %s", err)
		}
	})
}

func TestUserModel_UpdatePassword(t *testing.T) {

	cfg := data.TestPostgresConfig()
	db, err := data.Open(cfg)
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}
	defer db.Close()
	userModel := &UserModel{DB: db}
	t.Run("Update Password happy path", func(t *testing.T) {
		userToInsert := User{
			Email:    "updatedpassword@localhost",
			Password: "veryinsecurepassword",
		}

		newError := userModel.Insert(&userToInsert)
		if newError != nil {
			t.Errorf("Expected no error, got %s", err)
		}

		// get user before password update to compare later
		user, newError := userModel.GetByEmail(userToInsert.Email)

		newError = userModel.UpdatePassword(int(userToInsert.ID), "newsecurepassword")
		if err != nil {
			t.Errorf("Expected no error, got %s", err)
		}

		updatedUser, err := userModel.GetByEmail(userToInsert.Email)
		if err != nil {
			t.Errorf("Expected no error, got %s", err)
		}

		if reflect.DeepEqual(user.Password, updatedUser.Password) {
			t.Errorf("Expected password to be updated")
		}
		err = userModel.DeleteUser(userToInsert.Email)
	})

	t.Run("User not found", func(t *testing.T) {
		err := userModel.UpdatePassword(999, "newpassword")
		if err == nil {
			t.Errorf("Expected error, got nil")
		}
	})
}

func TestUserModel_DeleteUser(t *testing.T) {
	cfg := data.TestPostgresConfig()
	db, err := data.Open(cfg)
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}
	defer db.Close()
	userModel := &UserModel{DB: db}

	t.Run("Delete User happy path", func(t *testing.T) {
		userToDelete := User{
			Email:    "deleteuser@localhost",
			Password: "veryinsecurepassword",
		}
		userModel.Insert(&userToDelete)
		err := userModel.DeleteUser(userToDelete.Email)
		if err != nil {
			t.Errorf("Expected no error, got %s", err)
		}
	})
	t.Run("User not found", func(t *testing.T) {
		userToDelete := User{
			Email:    "deleteuser@localhost",
			Password: "veryinsecurepassword",
		}
		err = userModel.Insert(&userToDelete)
		if err != nil {
			t.Errorf("Expected no error, got %s", err)
		}
		err = userModel.DeleteUser(userToDelete.Email)
		if err != nil {
			t.Errorf("Expected no error, got %s", err)
		}
		err = userModel.DeleteUser(userToDelete.Email)
		if err == nil {
			t.Errorf("Expected error, got nil")
		}
	})
}

func TestUserModel_Authenticate(t *testing.T) {
	cfg := data.TestPostgresConfig()
	db, err := data.Open(cfg)
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}
	defer db.Close()
	userModel := &UserModel{DB: db}

	t.Run("Authenticate happy path", func(t *testing.T) {
		userToDelete := User{
			Email:    "authenticateuser@localhost",
			Password: "veryinsecurepassword",
		}
		passwwordBeforeHash := userToDelete.Password
		userModel.Insert(&userToDelete)
		user, err := userModel.Authenticate(userToDelete.Email, passwwordBeforeHash)
		if err != nil {
			t.Errorf("Expected no error, got %s", err)
		}
		if user.Email != userToDelete.Email {
			t.Errorf("Expected user to be returned")
		}
		userModel.DeleteUser(userToDelete.Email)

	})
	t.Run("invalid password", func(t *testing.T) {
		userToDelete := User{
			Email:    "authenticateuser@localhost",
			Password: "veryinsecurepassword",
		}
		err := userModel.Insert(&userToDelete)
		user, err := userModel.Authenticate(userToDelete.Email, "invalidpassword")
		if err == nil {
			t.Errorf("Expected error, got nil")
		}
		if user != nil {
			t.Errorf("Expected nil, got user")
		}
		err = userModel.DeleteUser(userToDelete.Email)
	})
	t.Run("User not found", func(t *testing.T) {
		user, err := userModel.Authenticate("notfound", "notfound")
		if err == nil {
			t.Errorf("Expected error, got nil")
		}
		if user != nil {
			t.Errorf("Expected nil, got user")
		}
	})

}

func TestUserModel_EnterPasswordHash(t *testing.T) {
	cfg := data.TestPostgresConfig()
	db, err := data.Open(cfg)
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}
	defer db.Close()
	userModel := &UserModel{DB: db}

	t.Run("EnterPasswordHash happy path", func(t *testing.T) {
		userToDelete := User{
			Email:    "deleteuser@localhost",
			Password: "veryinsecurepassword",
		}
		err := userModel.Insert(&userToDelete)
		if err != nil {
			t.Errorf("Expected no error, got %s", err)
		}

		passwordToken, salt, err := token.GenerateTokenAndSalt(32, 16)
		if err != nil {
			t.Errorf("Expected no error, got %s", err)
		}

		hashedToken := token.HashToken(passwordToken, salt)
		err = userModel.EnterPasswordHash(userToDelete.Email, string(hashedToken), string(salt))
		if err != nil {
			t.Errorf("Expected no error, got %s", err)
		}
		user, err := userModel.GetByEmail(userToDelete.Email)
		if err != nil {
			t.Errorf("Expected no error, got %s", err)
		}
		t.Log(user)
		if user.PasswordResetHashToken != hashedToken {
			t.Errorf("got %s, want %s", user.PasswordResetHashToken, hashedToken)
		}
		if user.PasswordResetSalt != string(salt) {
			t.Errorf("got %s, want %s", user.PasswordResetSalt, salt)
		}
		err = userModel.DeleteUser(userToDelete.Email)
		if err != nil {
			t.Errorf("Expected no error, got %s", err)
		}
	})
	t.Run("User not found", func(t *testing.T) {
		passwordToken, salt, err := token.GenerateTokenAndSalt(32, 16)
		if err != nil {
			t.Errorf("Expected no error, got %s", err)
		}
		err = userModel.EnterPasswordHash("notfound", string(passwordToken), string(salt))
		if err == nil {
			t.Errorf("Expected error, got nil")
		}
	})
}

func TestUserModel_ConsumePasswordReset(t *testing.T) {
	cfg := data.TestPostgresConfig()
	db, err := data.Open(cfg)
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}
	defer db.Close()
	userModel := &UserModel{DB: db}
	t.Run("Happy path", func(t *testing.T) {
		userToDelete := User{
			Email:    "deleteuser@localhost",
			Password: "veryinsecurepassword",
		}
		err := userModel.Insert(&userToDelete)
		if err != nil {
			t.Errorf("Expected no error, got %s", err)
		}

		passwordToken, salt, err := token.GenerateTokenAndSalt(32, 16)
		if err != nil {
			t.Errorf("Expected no error, got %s", err)
		}

		hashedToken := token.HashToken(passwordToken, salt)
		err = userModel.EnterPasswordHash(userToDelete.Email, string(hashedToken), string(salt))
		if err != nil {
			t.Errorf("Expected no error, got %s", err)
		}
		err = userModel.ConsumePasswordReset(userToDelete.Email)
		if err != nil {
			t.Errorf("Expected no error, got %s", err)
		}
		user, err := userModel.GetByEmail(userToDelete.Email)
		if err != nil {
			t.Errorf("Expected no error, got %s", err)
		}
		if !user.PasswordResetExpiry.IsZero() {
			t.Errorf("Expected PasswordResetExpiry to be unset, got %s", user.PasswordResetExpiry)
		}
		if user.PasswordResetHashToken != "" || user.PasswordResetSalt != "" {
			t.Errorf("Expected PasswordResetToken and PasswordResetSalt to be unset, got %s and %s", user.PasswordResetHashToken, user.PasswordResetSalt)
		}
		err = userModel.DeleteUser(userToDelete.Email)
		if err != nil {
			t.Errorf("Expected no error, got %s", err)
		}
	})
	t.Run("User not found", func(t *testing.T) {
		err = userModel.ConsumePasswordReset("notfound")
		if err == nil {
			t.Errorf("Expected error, got nil")
		}
	})
}
