package models

import (
	"reflect"
	"testing"
	"the_lonely_road/data"
	"time"
)

type App struct {
	userModel *UserModel
}

func TestUserModel_Insert(t *testing.T) {
	mockUser := User{
		ID:        1,
		Email:     "justatest@test.com",
		Password:  "mockpassword",
		CreatedAt: time.Now(),
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
			ID:       1,
			Email:    "admin@localhost",
			Password: "$2a$10$m2RvoCSnhAMGZggN1SPPsOwlSC8Ne0EX.wi7EHK2/pKKmoOmDQsUe",
		}

		user, err := userModel.GetByEmail(userToInsert.Email)
		if err != nil {
			t.Errorf("Expected no error, got %s", err)
		}

		if reflect.DeepEqual(user, &userToInsert) {
			t.Errorf("Expected user to be returned")
		}
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
