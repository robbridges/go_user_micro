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
	user, err := userModel.GetByEmail(mockUser.Email)
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}
	if !reflect.DeepEqual(user, &mockUser) {
		t.Errorf("Expected user to be returned")
	}
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
}
