package models

import (
	"reflect"
	"testing"
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
