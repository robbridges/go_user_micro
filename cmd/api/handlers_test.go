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
}

func TestApp_CreateUser_SadPaths(t *testing.T) {
	tests := []struct {
		name          string
		payload       []byte
		expectedCode  int
		expectedError string
	}{
		{
			name:          "Bad json",
			payload:       badPayload,
			expectedCode:  http.StatusBadRequest,
			expectedError: "body contains badly-form JSON (at character 30)\n",
		},
		{
			name:          "Duplicate user",
			payload:       payload,
			expectedCode:  http.StatusInternalServerError,
			expectedError: "duplicate email\n",
		},
		{
			name:          "Bad email",
			payload:       badEmailPayload,
			expectedCode:  http.StatusBadRequest,
			expectedError: "User password must be 4 characters long and email must be 5 characters long\n",
		},
	}
	app := App{userModel: &models.UserModelMock{DB: []*models.User{}}}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			if strings.Contains(test.name, "Duplicate user") {
				// Add the user before the test if the test name contains "Duplicate user."
				req, err := http.NewRequest("POST", "/users", bytes.NewBuffer(payload))
				if err != nil {
					t.Errorf("Unexpected error in creating HTTP request: %v", err)
				}
				rr := httptest.NewRecorder()
				app.CreateUser(rr, req)
			}

			req, err := http.NewRequest("POST", "/users", bytes.NewBuffer(test.payload))
			if err != nil {
				t.Errorf("Unexpected error in creating HTTP request: %v", err)
			}

			rr := httptest.NewRecorder()
			app.CreateUser(rr, req)

			if rr.Code != test.expectedCode {
				t.Errorf("Expected status code %d, got %d", test.expectedCode, rr.Code)
			}

			if test.name == "Bad json" || test.name == "Bad email" {
				var response models.User
				err = json.Unmarshal(rr.Body.Bytes(), &response)
				if err == nil {
					t.Errorf("Error expected when unmarshaling JSON: %v", err)
				}
			}

			if test.expectedError != "" && rr.Body.String() != test.expectedError {
				t.Errorf("Expected '%s', but got '%s'", test.expectedError, rr.Body.String())
			}

			// Check the mock DB size if the test case name is "Duplicate user."
			if test.name == "Duplicate user" {
				app.checkMockDBSize(t, 1)
				err = app.userModel.DeleteUser("test@example.com")
				if err != nil {
					t.Errorf("Unexpected error in deleting user: %v", err)
				}
			} else {
				app.checkMockDBSize(t, 0)
			}

		})
	}
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

		testPayload := []byte(`{"email": "test@example.com"}`)
		req, err := http.NewRequest("GET", "/users", bytes.NewBuffer(testPayload))
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
}

func TestGetUserByEmail_SadPaths(t *testing.T) {
	tests := []struct {
		name          string
		payload       []byte
		expectedCode  int
		expectedError string
	}{
		{
			name:          "Bad json",
			payload:       []byte(`{"email: "badjson"}`),
			expectedCode:  http.StatusBadRequest,
			expectedError: "body contains badly-form JSON (at character 11)\n",
		},
		{
			name:          "User not found",
			payload:       []byte(`{"email": "test2@example.com"}`),
			expectedCode:  http.StatusBadRequest,
			expectedError: "record not found\n",
		},
		{
			name:          "Bad email req",
			payload:       emailOnlyBadPayload,
			expectedCode:  http.StatusBadRequest,
			expectedError: "User password must be 4 characters long and email must be 5 characters long\n",
		},
	}

	app := App{userModel: &models.UserModelMock{DB: []*models.User{}}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/users", bytes.NewBuffer(test.payload))
			if err != nil {
				t.Errorf("Unexpected error in get request to /users")
			}

			rr := httptest.NewRecorder()
			app.getUserByEmail(rr, req)

			if rr.Code != test.expectedCode {
				t.Errorf("Expected status code %d, got %d", test.expectedCode, rr.Code)
			}

			if test.name == "Bad json" || test.name == "User not found" {
				var secondResponseUser models.User
				err = json.Unmarshal(rr.Body.Bytes(), &secondResponseUser)
				if err == nil {
					t.Error(test.expectedError)
				}
			}

			if rr.Body.String() != test.expectedError {
				t.Errorf("Expected '%s', got '%s'", test.expectedError, rr.Body.String())
			}
		})
	}
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
}

func TestApp_updateUserPassword_SadPaths(t *testing.T) {
	testCases := []struct {
		name             string
		payload          []byte
		expectedCode     int
		expectedResponse string
	}{
		{
			name:             "Bad JSON",
			payload:          []byte(`{"email": "test@example.com" "password": "securepassword"}`),
			expectedCode:     http.StatusBadRequest,
			expectedResponse: "body contains badly-form JSON (at character 30)",
		},
		{
			name:             "User not found",
			payload:          []byte(`{"email": "test2@example.com"}`),
			expectedCode:     http.StatusBadRequest,
			expectedResponse: "record not found",
		},
		{
			name:             "Bad email request",
			payload:          emailOnlyBadPayload,
			expectedCode:     http.StatusBadRequest,
			expectedResponse: "User password must be 4 characters long and email must be 5 characters long",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			app := App{userModel: &models.UserModelMock{DB: []*models.User{{ID: 1}}}}

			req, err := http.NewRequest("PATCH", "/users", bytes.NewBuffer(tc.payload))
			if err != nil {
				t.Errorf("Unexpected error in creating HTTP request: %v", err)
			}

			rr := httptest.NewRecorder()
			app.updateUserPassword(rr, req)

			if rr.Code != tc.expectedCode {
				t.Errorf("Expected status code %d, got %d", tc.expectedCode, rr.Code)
			}

			if !strings.Contains(rr.Body.String(), tc.expectedResponse) {
				t.Errorf("Expected body to contain '%s', but got '%s'", tc.expectedResponse, rr.Body.String())
			}
		})
	}
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
}

func TestProcessPasswordReset_SadPaths(t *testing.T) {
	tests := []struct {
		name          string
		token         string
		payload       []byte
		expectedCode  int
		expectedError string
	}{
		{
			name:          "Bad json",
			token:         "valid_token",
			payload:       []byte("badPayload"),
			expectedCode:  http.StatusBadRequest,
			expectedError: "body contains badly-form JSON",
		},
		{
			name:          "User not found",
			token:         "valid_token",
			payload:       []byte(`{"email": "notfound", "password": "securepassword"}`),
			expectedCode:  http.StatusBadRequest,
			expectedError: "record not found",
		},
		{
			name:          "Expired token",
			token:         "expired_token",
			payload:       []byte(`{"email": "test@example.com", "password": "securepassword"}`),
			expectedCode:  http.StatusBadRequest,
			expectedError: "Password reset token has expired",
		},
		{
			name:          "Bad token",
			token:         "invalid_token",
			payload:       []byte(`{"email": "test@example.com", "password": "securepassword"}`),
			expectedCode:  http.StatusBadRequest,
			expectedError: "Invalid Token",
		},
	}

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

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Hash and salt the password here
			passwordToken, salt, err := token.GenerateTokenAndSalt(32, 16)
			if err != nil {
				t.Errorf("Unexpected error in hashing token")
			}
			hashedToken := token.HashToken(passwordToken, salt)
			err = mockModel.EnterPasswordHash(user.Email, hashedToken, salt)
			if err != nil {
				t.Errorf("Unexpected error in entering password hash")
			}

			// Append "bad" to the token for the "Bad token" test case
			if test.name == "Bad token" {
				test.token += "bad"
			}
			// expire the password token if expired token test
			if test.name == "Expired token" {
				user.PasswordResetExpiry = time.Now().Add(-time.Hour)
			}

			// Create a recorder and request
			rr := httptest.NewRecorder()
			req, err := http.NewRequest("POST", "/users/password/reset?token="+test.token, bytes.NewBuffer(test.payload))
			if err != nil {
				t.Errorf("Unexpected error in creating HTTP request: %v", err)
			}

			// Call the ProcessPasswordReset function with the test case.
			app.ProcessPasswordReset(rr, req)

			if rr.Code != test.expectedCode {
				t.Errorf("Expected status code %d, got %d", test.expectedCode, rr.Code)
			}

			responseBody := rr.Body.String()
			if !strings.Contains(responseBody, test.expectedError) {
				t.Errorf("Expected body to contain '%s', but got '%s'", test.expectedError, responseBody)
			}
		})
	}
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
		testPayload := []byte(`{"email": "admin@admin.com", "password": "admin"}`)
		req, err := http.NewRequest("POST", "/users/login", bytes.NewBuffer(testPayload))
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
}

func TestApp_Authenticate_SadPaths(t *testing.T) {
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

	testCases := []struct {
		name             string
		payload          []byte
		expectedCode     int
		expectedResponse string
	}{
		{
			name:             "Bad email",
			payload:          []byte(`{"email": "abc", "password": "securepassword"}`),
			expectedCode:     http.StatusBadRequest,
			expectedResponse: "User password must be 4 characters long and email must be 5 characters long\n",
		},
		{
			name:             "Bad JSON",
			payload:          badPayload,
			expectedCode:     http.StatusBadRequest,
			expectedResponse: "body contains badly-form JSON (at character 30)\n",
		},
		{
			name:             "User not found",
			payload:          []byte(`{"email": "notfound", "password": "admin"}`),
			expectedCode:     http.StatusBadRequest,
			expectedResponse: "Invalid Credentials\n",
		},
		{
			name:             "Bad password",
			payload:          []byte(`{"email": "admin@admin.com", "password": "badpassword"}`),
			expectedCode:     http.StatusBadRequest,
			expectedResponse: "Invalid Credentials\n",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			req, err := http.NewRequest("POST", "/users/login", bytes.NewBuffer(testCase.payload))
			if err != nil {
				t.Errorf("Unexpected error in POST request to /users/login: %v", err)
			}
			rr := httptest.NewRecorder()
			app.Authenticate(rr, req)
			if rr.Code != testCase.expectedCode {
				t.Errorf("Expected status code %d, got %d", testCase.expectedCode, rr.Code)
			}
			if rr.Body.String() != testCase.expectedResponse {
				t.Errorf("Expected response '%s', got '%s'", testCase.expectedResponse, rr.Body.String())
			}
		})
	}
}

func TestApp_SignOut(t *testing.T) {
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
		testPayload := []byte(`{"email": "admin@admin.com", "password": "admin"}`)
		req, err := http.NewRequest("POST", "/users/login", bytes.NewBuffer(testPayload))
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
		req, err = http.NewRequest("POST", "/users/logout", nil)
		if err != nil {
			t.Errorf("Unexpected error in POST request to /users/logout")
		}

		req.AddCookie(rr.Result().Cookies()[0])
		rr = httptest.NewRecorder()
		app.SignOut(rr, req)
		if rr.Code != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, rr.Code)
		}
		cookie, err := req.Cookie("auth_token")
		if err != nil {
			t.Errorf("Unexpected error in getting cookie")
		}
		if !cookie.Expires.Before(time.Now()) {
			t.Errorf("Expected cookie to be expired, but got %s", cookie.Expires)
		}

	})
	t.Run("Sad Path", func(t *testing.T) {

		testPayload := []byte(`{"email": "admin@admin.com", "password": "admin"}`)
		req, err := http.NewRequest("POST", "/users/login", bytes.NewBuffer(testPayload))
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
		req, err = http.NewRequest("POST", "/users/logout", nil)
		if err != nil {
			t.Errorf("Unexpected error in POST request to /users/logout")
		}

		rr = httptest.NewRecorder()
		app.SignOut(rr, req)
		if rr.Code != http.StatusUnauthorized {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, rr.Code)
		}
		if rr.Body.String() != "Unauthorized\n" {
			t.Errorf("Expected response 'Unauthorized', got '%s'", rr.Body.String())
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
