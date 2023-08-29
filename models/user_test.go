package models

import (
	"reflect"
	"testing"
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
	err := userModel.Insert(&mockUser)
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}

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
	err := userModel.Insert(&mockUser)
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}
	t.Run("Happy path", func(t *testing.T) {
		err = userModel.UpdatePassword(int(mockUser.ID), "newpassword")
		if err != nil {
			t.Errorf("Expected no error, got %s", err)
		}
		user, err := userModel.GetByEmail(mockUser.Email)
		if err != nil {
			t.Errorf("Expected no error, got %s", err)
		}
		if user.Password != "newpassword" {
			t.Errorf("Expected password to be updated")
		}
	})
	t.Run("User not found", func(t *testing.T) {
		err = userModel.UpdatePassword(999, "newpassword")
		if err == nil && err.Error() != "user not found" {
			t.Errorf("Expected error, got %s", err)
		}
	})
}
