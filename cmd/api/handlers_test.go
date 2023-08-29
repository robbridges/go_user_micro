package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"the_lonely_road/models"
)

var payload = []byte(`{"email": "test@example.com", "password": "securepassword"}`)
var badPayload = []byte(`{"email": "test@example.com" "password": "securepassword"}`)
var jsonError = "Wrong Json Marshalled"

func TestApp_HandleHome(t *testing.T) {
	app := App{}

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Errorf("Unexpected error in get request to %s", req.URL)
	}

	rr := httptest.NewRecorder()

	app.HandleHome(rr, req)

	want := http.StatusOK
	got := rr.Code

	if got != want {
		t.Errorf("Got %d, want %d", got, want)
	}

	var response jsonPayload
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Error unmarshaling JSON: %v", err)
	}

	if response.Name != "User greet" {
		t.Error(jsonError)
	}

	if response.Data != "Hello user" {
		t.Error(jsonError)
	}

}

func TestApp_CreateUser(t *testing.T) {
	t.Run("Good json happy path", func(t *testing.T) {
		app := App{userModel: &models.UserModelMock{DB: []*models.User{}}}

		req, err := http.NewRequest("POST", "/users", bytes.NewBuffer(payload))
		if err != nil {
			t.Errorf("Unexpected error in get request to %s", req.URL)
		}

		rr := httptest.NewRecorder()

		app.CreateUser(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, rr.Code)
		}
		var response models.User
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		if err != nil {
			t.Errorf("Error unmarshaling JSON: %v", err)
		}
		// we aren't setting the id in the handler it's scanned by postgres so the id will always be 0
		if response.ID != 0 {
			t.Errorf("Expected ID to be 0, got %d", response.ID)
		}
		if response.Email != "test@example.com" {
			t.Errorf("Expected email to be test@example.com, but got %s", response.Email)
		}
		// the password should be omitted from the responses
		if response.Password != "" {
			t.Errorf("Expected password to be encrypted, but got %s", response.Password)
		}

		// cast user model to mock to check array length
		app.checkMockDBSize(t, 1)
	})
	t.Run("Bad json", func(t *testing.T) {
		app := App{userModel: &models.UserModelMock{DB: []*models.User{}}}

		req, err := http.NewRequest("POST", "/users", bytes.NewBuffer(badPayload))
		if err != nil {
			t.Errorf("Unexpected error in get request to %s", req.URL)
		}

		rr := httptest.NewRecorder()

		app.CreateUser(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, rr.Code)
		}
		var response models.User
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		if err == nil {
			t.Errorf("Error expected when unmarshaling JSON: %v", err)
		}
		app.checkMockDBSize(t, 0)
	})
	t.Run("Duplicate user", func(t *testing.T) {
		app := App{userModel: &models.UserModelMock{DB: []*models.User{}}}

		req, err := http.NewRequest("POST", "/users", bytes.NewBuffer(payload))
		if err != nil {
			t.Errorf("Unexpected error in get request to %s", req.URL)
		}

		rr := httptest.NewRecorder()

		app.CreateUser(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, rr.Code)
		}

		var response models.User
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		if err != nil {
			t.Errorf("Error expected when unmarshaling JSON: %v", err)
		}

		req, err = http.NewRequest("POST", "/users", bytes.NewBuffer(payload))
		rr = httptest.NewRecorder()
		app.CreateUser(rr, req)
		if rr.Code != http.StatusInternalServerError {
			t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, rr.Code)
		}
		app.checkMockDBSize(t, 1)
	})
}

func TestApp_getUserByEmail(t *testing.T) {
	app := App{userModel: &models.UserModelMock{DB: []*models.User{}}}

	req, err := http.NewRequest("POST", "/users", bytes.NewBuffer(payload))
	if err != nil {
		t.Errorf("Unexpected error in get request to /")
	}

	rr := httptest.NewRecorder()

	app.CreateUser(rr, req)

	payload = []byte(`{"email": "test@example.com"}`)
	req, err = http.NewRequest("GET", "/users", bytes.NewBuffer(payload))
	if err != nil {
		t.Errorf("Unexpected error in get request to /users")
	}

	var responseUser models.User
	err = json.Unmarshal(rr.Body.Bytes(), &responseUser)
	if err != nil {
		t.Errorf("Error unmarshaling JSON: %v", err)
	}
	rr = httptest.NewRecorder()
	app.getUserByEmail(rr, req)
	var secondResponseUser models.User
	err = json.Unmarshal(rr.Body.Bytes(), &secondResponseUser)

	if err != nil {
		t.Errorf("Error unmarshaling JSON: %v", err)
	}

	if responseUser != secondResponseUser {
		t.Errorf("Expected the same user to be returned but got %v and %v", responseUser, secondResponseUser)
	}

	app.checkMockDBSize(t, 1)
}

func TestApp_updateUserPassword(t *testing.T) {
	app := App{userModel: &models.UserModelMock{DB: []*models.User{}}}

	payload := []byte(`{"email": "test@example.com", "password": "securepassword"}`)

	req, err := http.NewRequest("POST", "/users", bytes.NewBuffer(payload))
	if err != nil {
		t.Errorf("Unexpected error in get request to %s", req.URL)
	}

	rr := httptest.NewRecorder()

	app.CreateUser(rr, req)

	var responseUser models.User
	err = json.Unmarshal(rr.Body.Bytes(), &responseUser)
	if err != nil {
		t.Errorf("Error unmarshaling JSON: %v", err)
	}

	req, err = http.NewRequest("PATCH", "/users", bytes.NewBuffer(payload))
	if err != nil {
		t.Errorf("Unexpected error in get request to /users")
	}
	rr = httptest.NewRecorder()

	app.updateUserPassword(rr, req)

	var secondResponseUser models.User
	err = json.Unmarshal(rr.Body.Bytes(), &secondResponseUser)
	if err != nil {
		t.Errorf("Error unmarshaling JSON: %v", err)
	}

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rr.Code)
	}

	// password not returned in the response, best we can do is check the code is 200 in integration we can actually
	//check the value in the db

	app.checkMockDBSize(t, 1)
}

func (app *App) checkMockDBSize(t *testing.T, expected int) {
	mockModel, ok := app.userModel.(*models.UserModelMock)
	if !ok {
		t.Errorf("Expected app.userModel to be of type UserModelMock")
	}
	if len(mockModel.DB) != expected {
		t.Errorf("Not enough users in database, expected %d, got %d", expected, len(mockModel.DB))
	}
}
