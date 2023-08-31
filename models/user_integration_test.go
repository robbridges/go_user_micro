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
	cfg := data.TestPostgresConfig()
	db, err := data.Open(cfg)
	defer db.Close()
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}
	mockUser := User{
		ID:        1,
		Email:     "justatest@test.com",
		Password:  "mockpassword",
		CreatedAt: time.Now(),
	}
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
}
