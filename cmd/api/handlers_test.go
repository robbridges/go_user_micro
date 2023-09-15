package main

import (
	"bytes"
	"encoding/json"
	"github.com/spf13/viper"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"the_lonely_road/errors"
	"the_lonely_road/mailer"
	"the_lonely_road/models"
	"the_lonely_road/token"
	"the_lonely_road/validator"
	"time"
)

// standard payloads for testing
var payload = []byte(`{"email": "test@example.com", "password": "securepassword"}`)
var badPayload = []byte(`{"email": "test@example.com" "password": "securepassword"}`)
var jsonError = "Wrong Json Marshalled"
var badEmailPayload = []byte(`{"email": "a", "password": "securepassword"}`)
var emailOnlyBadPayload = []byte(`{"email": "a"}`)
var badPasswordGoodEmailPayload = []byte(`{"email": "test@example.com", "password": "a"}`)

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
		v := validator.New()
		rr := httptest.NewRecorder()

		app.CreateUser(rr, req)
		if !v.Valid() {
			t.Errorf("Expected validator to be valid, got %v", v.Errors)
		}
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
			t.Errorf("Expected email to be test@example.com, got %s", response.Email)
		}
		// the password should be omitted from the responses
		if response.Password == "securepassword" {
			t.Errorf("Expected password to be encrypted, but got %s", response.Password)
		}
		if rr.Header().Get("Set-Cookie") == "" {
			t.Errorf("Expected cookie to be set, but got %s", rr.Header().Get("Set-Cookie"))
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
		// no user should be entered as the function errors and returns
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
	t.Run("Bad email", func(t *testing.T) {
		app := App{userModel: &models.UserModelMock{DB: []*models.User{}}}

		req, err := http.NewRequest("POST", "/users", bytes.NewBuffer(badEmailPayload))
		if err != nil {
			t.Errorf("Unexpected error in get request to %s", req.URL)
		}

		rr := httptest.NewRecorder()

		app.CreateUser(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, rr.Code)
		}

		// validate that the validator did not like the user
		if string(rr.Body.String()) != "User password must be 4 characters long and email must be 5 characters long\n" {
			t.Errorf("Expected bad user error, got %s", rr.Body.String())
		}
		// no user should be implemented
		app.checkMockDBSize(t, 0)
	})

}

func TestApp_getUserByEmail(t *testing.T) {
	app := App{userModel: &models.UserModelMock{DB: []*models.User{}}}
	user := models.User{
		ID:        1,
		Password:  "secret",
		Email:     "test@example.com",
		CreatedAt: time.Now(),
	}

	mockModel, ok := app.userModel.(*models.UserModelMock)
	if !ok {
		t.Errorf("Expected app.userModel to be of type UserModelMock")
	}
	mockModel.DB = append(mockModel.DB, &user)
	t.Run("Happy Path", func(t *testing.T) {

		payload = []byte(`{"email": "test@example.com"}`)
		req, err := http.NewRequest("GET", "/users", bytes.NewBuffer(payload))
		if err != nil {
			t.Errorf("Unexpected error in get request to /users")
		}

		rr := httptest.NewRecorder()
		if rr.Code != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, rr.Code)
		}
		app.getUserByEmail(rr, req)
		var responseUser models.User
		err = json.Unmarshal(rr.Body.Bytes(), &responseUser)

		if err != nil {
			t.Errorf("Error unmarshaling JSON: %v", err)
		}

		if reflect.DeepEqual(responseUser, user) {
			t.Errorf("Expected user to be returned")
		}
		app.checkMockDBSize(t, 1)
	})
	t.Run("Bad json", func(t *testing.T) {

		badEmailpayload := []byte(`{"email: "badjson"}`)
		req, err := http.NewRequest("GET", "/users", bytes.NewBuffer(badEmailpayload))
		if err != nil {
			t.Errorf("Unexpected error in get request to /users")
		}
		rr := httptest.NewRecorder()
		app.getUserByEmail(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, rr.Code)
		}
		var secondResponseUser models.User
		err = json.Unmarshal(rr.Body.Bytes(), &secondResponseUser)
		if err == nil {
			t.Error("bad json should have thrown error but got nil")
		}
	})
	t.Run("User not found", func(t *testing.T) {

		payload = []byte(`{"email": "test2@example.com"}`)
		req, err := http.NewRequest("GET", "/users", bytes.NewBuffer(payload))
		if err != nil {
			t.Errorf("Unexpected error in get request to /users")
		}
		rr := httptest.NewRecorder()
		app.getUserByEmail(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, rr.Code)
		}
		var secondResponseUser models.User
		err = json.Unmarshal(rr.Body.Bytes(), &secondResponseUser)
		if err == nil {
			t.Error("no data to return should result in error")
		}

		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, rr.Code)
		}

		if rr.Body.String() != "record not found\n" {
			t.Errorf("Expected record not found error, got %s", rr.Body.String())
		}
	})
	t.Run("Bad email req", func(t *testing.T) {
		app := App{userModel: &models.UserModelMock{DB: []*models.User{}}}

		req, err := http.NewRequest("GET", "/users", bytes.NewBuffer(emailOnlyBadPayload))
		if err != nil {
			t.Errorf("Unexpected error in get request to %s", req.URL)
		}

		rr := httptest.NewRecorder()

		app.getUserByEmail(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, rr.Code)
		}
		if string(rr.Body.String()) != "User password must be 4 characters long and email must be 5 characters long\n" {
			t.Errorf("Expected bad user error, got %s", rr.Body.String())
		}
	})
}

func TestApp_updateUserPassword(t *testing.T) {
	viper.SetConfigFile("../../email.env")
	if err := viper.ReadInConfig(); err != nil {
		t.Fatalf("failed to read config file: %v", err)
	}
	emailCfg := mailer.DefaultSMTPConfig()
	mailClient := mailer.NewEmailService(emailCfg)

	app := App{userModel: &models.UserModelMock{DB: []*models.User{}}, emailer: mailClient}
	user := models.User{
		ID:        1,
		Password:  "secret",
		Email:     "test@example.com",
		CreatedAt: time.Now(),
	}

	mockModel, ok := app.userModel.(*models.UserModelMock)
	if !ok {
		t.Errorf("Expected app.userModel to be of type UserModelMock")
	}
	mockModel.DB = append(mockModel.DB, &user)
	t.Run("Happy Path", func(t *testing.T) {

		emailPayload := []byte(`{"email": "test@example.com"}`)
		req, err := http.NewRequest("PATCH", "/users", bytes.NewBuffer(emailPayload))
		if err != nil {
			t.Errorf("Unexpected error in get request to /users")
		}
		rr := httptest.NewRecorder()

		app.updateUserPassword(rr, req)

		var resp string
		err = json.Unmarshal(rr.Body.Bytes(), &resp)
		if err != nil {
			t.Errorf("Error unmarshaling JSON: %v", err)
		}

		if rr.Code != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, rr.Code)
		}

		if err != nil {
			t.Errorf("Unexpected error reading response body: %v", err)
		}
		if user.PasswordResetSalt == "" || user.PasswordResetHashToken == "" {
			t.Errorf("Expected user password salts and hash to be set")
		}

		want := errors.PasswordResetEmail
		if !strings.Contains(resp, want) {
			t.Errorf("Expected body %s, but got %s", want, resp)
		}
	})
	t.Run("Bad json", func(t *testing.T) {

		rr := httptest.NewRecorder()
		payload = []byte(`{"email": "test@example.com" "password": "securepassword"}`)
		req, err := http.NewRequest("PATCH", "/users", bytes.NewBuffer(payload))
		if err != nil {
			t.Errorf("Unexpected error in patch request to /users")
		}
		app.updateUserPassword(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, rr.Code)
		}
	})
	t.Run("User not found", func(t *testing.T) {

		payload = []byte(`{"email": "test2@example.com"}`)
		req, err := http.NewRequest("PATCH", "/users", bytes.NewBuffer(payload))
		if err != nil {
			t.Errorf("Unexpected error in get request to /users")
		}
		rr := httptest.NewRecorder()

		app.updateUserPassword(rr, req)

		t.Log(rr.Body.String())
		if string(rr.Body.String()) != "record not found\n" {
			t.Errorf("Expected record not found error, got %s", rr.Body.String())
		}

		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, rr.Code)
		}

		app.checkMockDBSize(t, 1)
	})
	t.Run("Bad email req", func(t *testing.T) {
		req, err := http.NewRequest("PATCH", "/users", bytes.NewBuffer(emailOnlyBadPayload))
		if err != nil {
			t.Errorf("Unexpected error in get request to %s", req.URL)
		}

		rr := httptest.NewRecorder()

		app.updateUserPassword(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, rr.Code)
		}
		if string(rr.Body.String()) != "User password must be 4 characters long and email must be 5 characters long\n" {
			t.Errorf("Expected bad user error, got %s", rr.Body.String())
		}
	})

}

func TestApp_ProcessPasswordReset(t *testing.T) {

	app := App{userModel: &models.UserModelMock{DB: []*models.User{}}}
	user := models.User{
		ID:        1,
		Password:  "secret",
		Email:     "test@example.com",
		CreatedAt: time.Now(),
	}

	mockModel, ok := app.userModel.(*models.UserModelMock)
	if !ok {
		t.Errorf("Expected app.userModel to be of type UserModelMock")
	}
	mockModel.DB = append(mockModel.DB, &user)
	t.Run("Happy Path", func(t *testing.T) {
		testUser, err := mockModel.GetByEmail(user.Email)
		if err != nil {
			t.Errorf("Unexpected error in get request to /users")
		}
		t.Log(testUser.PasswordResetExpiry)

		passwordToken, salt, err := token.GenerateTokenAndSalt(32, 16)
		if err != nil {
			t.Errorf("Unexpected error in get request to /users")
		}
		hashedToken := token.HashToken(passwordToken, salt)
		err = mockModel.EnterPasswordHash(testUser.Email, hashedToken, salt)
		if err != nil {
			t.Errorf("Unexpected error in get request to /users")
		}
		if testUser.PasswordResetExpiry != user.PasswordResetExpiry {
			t.Errorf("Expected PasswordResetExpiry to be %s, got %s", user.PasswordResetExpiry, testUser.PasswordResetExpiry)
		}

		testPayload := []byte(`{"email": "test@example.com", "password": "securepassword"}`)
		req, err := http.NewRequest("POST", "/users/password/reset?token="+passwordToken, bytes.NewBuffer(testPayload))
		if err != nil {
			t.Errorf("Unexpected error in get request to /users")
		}
		t.Log(testUser.PasswordResetExpiry)
		rr := httptest.NewRecorder()
		app.ProcessPasswordReset(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, rr.Code)
		}
		want := rr.Body.String()
		if !strings.Contains(want, "Password updated successfully") {
			t.Errorf("Expected body %s, but got %s", "password updated successfully", rr.Body.String())
		}
	})
	t.Run("Bad json", func(t *testing.T) {
		passwordToken, _, err := token.GenerateTokenAndSalt(32, 16)
		if err != nil {
			t.Errorf("Unexpected error in get request to /users")
		}

		rr := httptest.NewRecorder()
		req, err := http.NewRequest("POST", "/users/password/reset?token="+passwordToken, bytes.NewBuffer(badPayload))
		if err != nil {
			t.Errorf("Unexpected error in patch request to /users")
		}
		app.ProcessPasswordReset(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, rr.Code)
		}
		got := rr.Body.String()
		if !strings.Contains(got, "body contains bady-form JSON") {
			t.Errorf("Expected body %s, but got %s", "body contains bady-form JSON", rr.Body.String())
		}
	})
	t.Run("User not found", func(t *testing.T) {
		passwordToken, _, err := token.GenerateTokenAndSalt(32, 16)
		if err != nil {
			t.Errorf("Unexpected error in hashing token")
		}
		rr := httptest.NewRecorder()
		payload = []byte(`{"email": "notfound", "password": "securepassword"}`)
		req, err := http.NewRequest("POST", "/users/password/reset?token="+passwordToken, bytes.NewBuffer(payload))
		if err != nil {
			t.Errorf("Unexpected error in POST request to /users")
		}
		app.ProcessPasswordReset(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, rr.Code)
		}
		got := rr.Body.String()
		if !strings.Contains(got, "record not found") {
			t.Errorf("Expected body %s, but got %s", "record not found", rr.Body.String())
		}
	})
	t.Run("Expired token", func(t *testing.T) {
		// Get the user from the mock model.
		newUser, err := mockModel.GetByEmail("test@example.com")
		if err != nil {
			t.Errorf("Failed to retrieve user: %v", err)
		}

		// Generate a password reset token, hash it, and set it in the mock model.
		passwordToken, salt, err := token.GenerateTokenAndSalt(32, 16)
		if err != nil {
			t.Errorf("Failed to generate token and salt: %v", err)
		}
		hashedToken := token.HashToken(passwordToken, salt)
		err = mockModel.EnterPasswordHash(newUser.Email, hashedToken, salt)
		if err != nil {
			t.Errorf("Failed to enter password hash: %v", err)
		}

		// Expire the token by setting PasswordResetExpiry to a time in the past.
		newUser.PasswordResetExpiry = time.Now().Add(-time.Hour)

		// Prepare a test payload and create an HTTP request.
		testPayload := []byte(`{"email": "test@example.com", "password": "securepassword"}`)
		req, err := http.NewRequest("POST", "/users/password/reset?token="+passwordToken, bytes.NewBuffer(testPayload))
		if err != nil {
			t.Errorf("Failed to create HTTP request: %v", err)
		}

		// Create a response recorder to capture the HTTP response.
		rr := httptest.NewRecorder()

		// Call the ProcessPasswordReset function with the expired token.
		app.ProcessPasswordReset(rr, req)

		// Check the response status code.
		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, rr.Code)
		}

		// Check the response body for the expected error message.
		expectedErrorMessage := errors.PasswordResetExpired
		responseBody := rr.Body.String()
		if !strings.Contains(responseBody, expectedErrorMessage) {
			t.Errorf("Expected body to contain '%s', but got '%s'", expectedErrorMessage, responseBody)
		}
	})
	t.Run("Bad token", func(t *testing.T) {
		user, err := mockModel.GetByEmail(user.Email)
		if err != nil {
			t.Errorf("Unexpected error in get request to /users")
		}

		passwordToken, salt, err := token.GenerateTokenAndSalt(32, 16)
		if err != nil {
			t.Errorf("Unexpected error in get request to /users")
		}
		hashedToken := token.HashToken(passwordToken, salt)
		err = mockModel.EnterPasswordHash(user.Email, hashedToken, salt)
		if err != nil {
			t.Errorf("Unexpected error in get request to /users")
		}
		testPayload := []byte(`{"email": "test@example.com", "password": "securepassword"}`)
		req, err := http.NewRequest("POST", "/users/password/reset?token="+passwordToken+"bad", bytes.NewBuffer(testPayload))
		if err != nil {
			t.Errorf("Unexpected error in get request to /users")
		}
		rr := httptest.NewRecorder()
		app.ProcessPasswordReset(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, rr.Code)
		}
		got := rr.Body.String()
		if !strings.Contains(got, errors.InvalidToken) {
			t.Errorf("Expected body %s, but got %s", errors.InvalidToken, rr.Body.String())
		}
	})
}

func TestApp_Authenticate(t *testing.T) {
	app := App{userModel: &models.UserModelMock{DB: []*models.User{}}}
	user := models.User{
		ID:        1,
		Password:  "admin",
		Email:     "admin@admin.com",
		CreatedAt: time.Now(),
	}

	mockModel, ok := app.userModel.(*models.UserModelMock)
	if !ok {
		t.Errorf("Expected app.userModel to be of type UserModelMock")
	}
	// actually insert the model instead of just appending to encrypt the password I should have done this everywhere.
	err := mockModel.Insert(&user)
	if err != nil {
		t.Errorf("Unexpected error in inserting user")
	}
	t.Run("Happy Path", func(t *testing.T) {
		payload = []byte(`{"email": "admin@admin.com", "password": "admin"}`)
		req, err := http.NewRequest("POST", "/users/login", bytes.NewBuffer(payload))
		if err != nil {
			t.Errorf("Unexpected error in POST request to /users/login")
		}
		rr := httptest.NewRecorder()

		app.Authenticate(rr, req)
		if rr.Code != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, rr.Code)
		}
		var responseUser models.User
		err = json.Unmarshal(rr.Body.Bytes(), &responseUser)
		if err != nil {
			t.Errorf("Error unmarshaling JSON: %v", err)
		}
		if responseUser.Email != user.Email {
			t.Errorf("Expected email to be %s, got %s", user.Email, responseUser.Email)
		}
		if rr.Header().Get("Set-Cookie") == "" {
			t.Errorf("Expected cookie to be set, but got %s", rr.Header().Get("Set-Cookie"))
		}
	})
	t.Run("Bad email", func(t *testing.T) {

		req, err := http.NewRequest("POST", "/users/login", bytes.NewBuffer(badEmailPayload))
		if err != nil {
			t.Errorf("Unexpected error in POST request to /users/login")
		}
		rr := httptest.NewRecorder()
		app.Authenticate(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, rr.Code)
		}
		if rr.Body.String() != "User password must be 4 characters long and email must be 5 characters long\n" {
			t.Errorf("Expected bad email error, got %s", rr.Body.String())
		}

	})
	t.Run("Bad json", func(t *testing.T) {
		req, err := http.NewRequest("POST", "/users/login", bytes.NewBuffer(badPayload))
		if err != nil {
			t.Errorf("Unexpected error in POST request to /users/login")
		}
		rr := httptest.NewRecorder()
		app.Authenticate(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, rr.Code)
		}
		if rr.Body.String() != "body contains bady-form JSON (at character 30)\n" {
			t.Errorf("Expected bad json error, got %s", rr.Body.String())
		}
	})
	t.Run("User not found", func(t *testing.T) {
		payload = []byte(`{"email": "notfound", "password": "admin"}`)
		req, err := http.NewRequest("POST", "/users/login", bytes.NewBuffer(payload))
		if err != nil {
			t.Errorf("Unexpected error in POST request to /users/login")
		}
		rr := httptest.NewRecorder()
		app.Authenticate(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, rr.Code)
		}
		if rr.Body.String() != "Invalid Credentials\n" {
			t.Errorf("Expected record not found error, got %s", rr.Body.String())
		}
	})
	t.Run("Bad password", func(t *testing.T) {
		payload = []byte(`{"email": "admin@admin.com", "password": "badpassword"}`)
		req, err := http.NewRequest("POST", "/users/login", bytes.NewBuffer(payload))
		if err != nil {
			t.Errorf("Unexpected error in POST request to /users/login")
		}
		rr := httptest.NewRecorder()
		app.Authenticate(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, rr.Code)
		}
		if rr.Body.String() != "Invalid Credentials\n" {
			t.Errorf("Expected record not found error, got %s", rr.Body.String())
		}
	})
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
