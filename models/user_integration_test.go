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

		err := userModel.Insert(&userToInsert)
		if err != nil {
			t.Errorf("Expected no error, got %s", err)
		}

		// get user before password update to compare later
		user, err := userModel.GetByEmail(userToInsert.Email)

		err = userModel.UpdatePassword(int(userToInsert.ID), "newsecurepassword")
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
		userModel.DeleteUser(userToInsert.Email)
	})

}
