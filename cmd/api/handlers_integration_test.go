package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/spf13/viper"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"the_lonely_road/data"
	"the_lonely_road/errors"
	"the_lonely_road/mailer"
	"the_lonely_road/models"
	"the_lonely_road/token"
	"time"
)

func TestApp_HandleHomeIntegration(t *testing.T) {
	testCfg := data.TestPostgresConfig()
	testDB, err := data.Open(testCfg)
	if err != nil {
		t.Errorf("Expected database to open, but got %s", err)
	}
	app := App{userModel: &models.UserModel{DB: testDB}}

	server := httptest.NewServer(http.HandlerFunc(app.HandleHome))
	defer server.Close()

	req, err := http.NewRequest("GET", server.URL, bytes.NewBuffer(payload))
	if err != nil {
		t.Fatal(err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d, but got %d", http.StatusOK, resp.StatusCode)
	}

	var response jsonPayload
	err = json.NewDecoder(resp.Body).Decode(&response)

	if response.Name != "User greet" {
		t.Error(jsonError)
	}

	if response.Data != "Hello user" {
		t.Error(jsonError)
	}
}

func TestApp_CreateUserIntegration(t *testing.T) {
	// get and apply db to app's user Model
	testCfg := data.TestPostgresConfig()
	testDB, err := data.Open(testCfg)
	defer testDB.Close()
	if err != nil {
		t.Errorf("Expected database to open, but got %s", err)
	}
	app := App{userModel: &models.UserModel{DB: testDB}}

	t.Run("Create user", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(app.CreateUser))
		defer server.Close()

		req, err := http.NewRequest("POST", server.URL+"/users", bytes.NewBuffer(payload))
		if err != nil {
			t.Fatal(err)
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status %d, but got %d", http.StatusOK, resp.StatusCode)
		}
		// parse user from response body
		var user models.User
		err = json.NewDecoder(resp.Body).Decode(&user)
		if err != nil {
			t.Errorf("Error unmarshaling JSON: %v", err)
		}
		expectedEmail := "test@example.com"

		if user.Email != expectedEmail {
			t.Errorf("Expected email %s, but got %s", expectedEmail, user.Email)
		}
		// Check that set-cookie is in the response. Since this will only be used a mock user
		//service we're okay just using stateless authentication
		if resp.Header.Get("Set-Cookie") == "" {
			t.Error("Expected cookie to be set, but got none")
		}
		// delete user to keep test db clean
		app.userModel.DeleteUser(user.Email)
	})
}

func TestApp_CreateUserIntegration_SadPaths(t *testing.T) {

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
			expectedCode:  http.StatusInternalServerError, // Update this as needed.
			expectedError: "duplicate email\n",
		},
		{
			name:          "Bad email",
			payload:       badEmailPayload,
			expectedCode:  http.StatusBadRequest,
			expectedError: "User password must be 4 characters long and email must be 5 characters long\n",
		},
	}

	testCfg := data.TestPostgresConfig()
	testDB, err := data.Open(testCfg)
	defer testDB.Close()
	if err != nil {
		t.Errorf("Expected database to open, but got %s", err)
	}
	app := App{userModel: &models.UserModel{DB: testDB}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Create a test HTTP server.
			server := httptest.NewServer(http.HandlerFunc(app.CreateUser))
			defer server.Close()

			// If the test name contains "Duplicate user," add the user before the test.
			if strings.Contains(test.name, "Duplicate user") {
				req, err := http.NewRequest("POST", server.URL+"/users", bytes.NewBuffer(payload))
				if err != nil {
					t.Fatalf("Unexpected error in creating HTTP request: %v", err)
				}

				// Make the HTTP request to add the user.
				resp, err := http.DefaultClient.Do(req)
				if err != nil {
					t.Fatalf("Unexpected error in making HTTP request: %v", err)
				}
				defer resp.Body.Close()

				// Check if the user was added successfully.
				if resp.StatusCode != http.StatusOK {
					t.Fatalf("Failed to add user for test '%s'", test.name)
				}
			}

			// Build the request with the test payload and server URL.
			req, err := http.NewRequest("POST", server.URL+"/users", bytes.NewBuffer(test.payload))
			if err != nil {
				t.Fatalf("Unexpected error in creating HTTP request: %v", err)
			}

			// Make the actual HTTP request for the test.
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("Unexpected error in making HTTP request: %v", err)
			}
			defer resp.Body.Close()

			// Check the HTTP response code and response body for expected results.
			if resp.StatusCode != test.expectedCode {
				t.Errorf("Expected status code %d, got %d", test.expectedCode, resp.StatusCode)
			}

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("Unexpected error reading response body: %v", err)
			}

			if test.name == "Bad json" || test.name == "Bad email" {
				var response models.User
				err = json.Unmarshal(body, &response)
				if err == nil {
					t.Errorf("Error expected when unmarshaling JSON: %v", err)
				}
			}

			if test.expectedError != "" && string(body) != test.expectedError {
				t.Errorf("Expected '%s', but got '%s'", test.expectedError, string(body))
			}

			if test.name == "Duplicate user" {
				// Delete the user if it was added for the test.
				app.userModel.DeleteUser("test@example.com")
			}
		})
	}
}

func TestApp_GetUserIntegration(t *testing.T) {
	testCfg := data.TestPostgresConfig()
	testDB, err := data.Open(testCfg)
	defer testDB.Close()
	if err != nil {
		t.Errorf("Expected database to open, but got %s", err)
	}
	app := App{userModel: &models.UserModel{DB: testDB}}
	t.Run("Get user Happy path", func(t *testing.T) {

		server := httptest.NewServer(http.HandlerFunc(app.getUserByEmail))
		defer server.Close()
		payloadEmailOnly := []byte(`{"email": "admin@localhost"}`)
		req, err := http.NewRequest("GET", server.URL+"/users)", bytes.NewBuffer(payloadEmailOnly))
		if err != nil {
			t.Errorf("Unexpected error in get request to %s", req.URL)
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status %d, but got %d", http.StatusOK, resp.StatusCode)
		}

		var userReturned models.User
		err = json.NewDecoder(resp.Body).Decode(&userReturned)
		if err != nil {
			t.Errorf("Error unmarshaling JSON: %v", err)
		}
		t.Log(userReturned)
		var expectedUser = models.User{
			ID:       1,
			Email:    "admin@localhost",
			Password: "$2a$10$m2RvoCSnhAMGZggN1SPPsOwlSC8Ne0EX.wi7EHK2/pKKmoOmDQsUe",
		}
		if reflect.DeepEqual(userReturned, expectedUser) {
			t.Errorf("Expected user %v, but got %v", expectedUser, userReturned)
		}
	})

	t.Run("Get user Sad path", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(app.getUserByEmail))
		defer server.Close()
		payload := []byte(`{"email": "adminx@localhost"}`)
		req, err := http.NewRequest("GET", server.URL+"/users)", bytes.NewBuffer(payload))
		if err != nil {
			t.Errorf("Unexpected error in get request to %s", req.URL)
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status %d, but got %d", http.StatusBadRequest, resp.StatusCode)
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Errorf("Unexpected error reading response body: %v", err)
		}

		expectedBody := "record not found\n"

		if string(body) != expectedBody {
			t.Errorf("Expected body %s, but got %s", expectedBody, string(body))
		}

	})
	t.Run("Get user Invalid email payload", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(app.getUserByEmail))
		defer server.Close()

		req, err := http.NewRequest("GET", server.URL+"/users", bytes.NewBuffer(emailOnlyBadPayload))
		if err != nil {
			t.Errorf("Unexpected error in get request to %s", req.URL)
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status %d, but got %d", http.StatusBadRequest, resp.StatusCode)
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Errorf("Unexpected error reading response body: %v", err)
		}

		if string(body) != "User password must be 4 characters long and email must be 5 characters long\n" {
			t.Errorf("Expected body %s, but got %s", "invalid user\n", string(body))
		}
	})
}

func Test_UpdatePasswordIntegration(t *testing.T) {
	viper.SetConfigFile("../../email.env")
	if err := viper.ReadInConfig(); err != nil {
		t.Fatalf("failed to read config file: %v", err)
	}
	testCfg := data.TestPostgresConfig()
	testDB, err := data.Open(testCfg)
	defer testDB.Close()
	if err != nil {
		t.Errorf("Expected database to open, but got %s", err)
	}
	mailCfg := mailer.DefaultSMTPConfig()
	mailClient := mailer.NewEmailService(mailCfg)
	app := App{userModel: &models.UserModel{DB: testDB}, emailer: mailClient}
	t.Run("Update password Happy path", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(app.updateUserPassword))
		defer server.Close()
		var emailPayload = []byte(`{"email": "admin@localhost"}`)
		req, err := http.NewRequest("PATCH", server.URL+"/users)", bytes.NewBuffer(emailPayload))
		if err != nil {
			t.Errorf("Unexpected error in get request to %s", req.URL)
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status %d, but got %d", http.StatusOK, resp.StatusCode)
		}
		user, err := app.userModel.GetByEmail("admin@localhost")
		if err != nil {
			t.Errorf("Expected no error, got %s", err)
		}
		if user.PasswordResetHashToken == "" || user.PasswordResetSalt == "" {
			t.Errorf("Expected salt and hash to be set")
		}

		if err != nil {
			t.Errorf("Unexpected error reading response body: %v", err)
		}
		response, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Errorf("Unexpected error reading response body: %v", err)
		}

		want := errors.PasswordResetEmail
		if !strings.Contains(string(response), want) {
			t.Errorf("Expected body %s, but got %s", want, string(response))
		}
	})
	t.Run("Update password Sad path", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(app.updateUserPassword))
		defer server.Close()
		emailPayload := []byte(`{"email": "adminx@localhost"}`)
		req, err := http.NewRequest("PATCH", server.URL+"/users)", bytes.NewBuffer(emailPayload))
		if err != nil {
			t.Errorf("Unexpected error in get request to %s", req.URL)
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status %d, but got %d", http.StatusBadRequest, resp.StatusCode)
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Errorf("Unexpected error reading response body: %v", err)
		}

		expectedBody := "record not found\n"

		if string(body) != expectedBody {
			t.Errorf("Expected body %s, but got %s", expectedBody, string(body))
		}
	})
	t.Run("Update password Invalid email payload", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(app.updateUserPassword))
		defer server.Close()

		req, err := http.NewRequest("PATCH", server.URL+"/users", bytes.NewBuffer(emailOnlyBadPayload))
		if err != nil {
			t.Errorf("Unexpected error in get request to %s", req.URL)
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status %d, but got %d", http.StatusBadRequest, resp.StatusCode)
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Errorf("Unexpected error reading response body: %v", err)
		}

		if string(body) != "User password must be 4 characters long and email must be 5 characters long\n" {
			t.Errorf("Expected body %s, but got %s", "invalid user\n", string(body))
		}
	})
}

func TestApp_ProcessPasswordResetIntegration(t *testing.T) {
	testCfg := data.TestPostgresConfig()
	testDB, err := data.Open(testCfg)
	defer testDB.Close()
	if err != nil {
		t.Errorf("Expected database to open, but got %s", err)
	}
	app := App{userModel: &models.UserModel{DB: testDB}}
	t.Run("Process password reset Happy path", func(t *testing.T) {
		hash, salt, err := token.GenerateTokenAndSalt(32, 16)
		if err != nil {
			t.Errorf("Expected no error, got %s", err)
		}
		hashedToken := token.HashToken(hash, salt)
		err = app.userModel.EnterPasswordHash("admin@localhost", hashedToken, salt)
		if err != nil {
			t.Errorf("Expected no error, got %s", err)
		}
		server := httptest.NewServer(http.HandlerFunc(app.ProcessPasswordReset))
		defer server.Close()
		var emailPayload = []byte(`{"email": "admin@localhost", "password": "newpassword"}`)
		req, err := http.NewRequest("POST", fmt.Sprintf(server.URL+"/users/password/reset?token=%s", hash), bytes.NewBuffer(emailPayload))
		if err != nil {
			t.Errorf("Unexpected error in get request to %s", req.URL)
		}
		response, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Errorf("Expected no error, got %s", err)
		}
		defer response.Body.Close()
		if response.StatusCode != http.StatusOK {
			t.Errorf("Expected status %d, but got %d", http.StatusOK, response.StatusCode)
		}
		want := "password updated succesfully"
		body, err := io.ReadAll(response.Body)
		if err != nil {
			t.Errorf("Unexpected error reading response body: %v", err)
		}
		if strings.Contains(string(body), want) {
			t.Errorf("Expected body %s, but got %s", want, string(body))
		}
		user, err := app.userModel.GetByEmail("admin@localhost")
		if err != nil {
			t.Errorf("Expected no error, got %s", err)
		}
		if user.PasswordResetHashToken != "" || user.PasswordResetSalt != "" {
			t.Errorf("Expected salt and hash to be empty")
		}
		if !user.PasswordResetExpiry.IsZero() {
			t.Errorf("Expected expiry to be empty")
		}
	})
	t.Run("Process password reset Sad path", func(t *testing.T) {
		hash, salt, err := token.GenerateTokenAndSalt(32, 16)
		if err != nil {
			t.Errorf("Expected no error, got %s", err)
		}
		hashedToken := token.HashToken(hash, salt)
		err = app.userModel.EnterPasswordHash("admin@localhost", hashedToken, salt)
		if err != nil {
			t.Errorf("Expected no error, got %s", err)
		}
		server := httptest.NewServer(http.HandlerFunc(app.ProcessPasswordReset))
		defer server.Close()
		var emailPayload = []byte(`{"email": "admin@localhost", "password": "newpassword"}`)
		req, err := http.NewRequest("POST", fmt.Sprintf(server.URL+"/users/password/reset?token=%swrong", hash), bytes.NewBuffer(emailPayload))
		if err != nil {
			t.Errorf("Unexpected error in get request to %s", req.URL)
		}
		response, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Errorf("Expected no error, got %s", err)
		}
		defer response.Body.Close()
		if response.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status %d, but got %d", http.StatusOK, response.StatusCode)
		}
		want := "invalid token\n"
		body, err := io.ReadAll(response.Body)
		if err != nil {
			t.Errorf("Unexpected error reading response body: %v", err)
		}
		if strings.Contains(string(body), want) {
			t.Errorf("Expected body %s, but got %s", want, string(body))
		}
	})
	t.Run("Process password reset Invalid payload", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(app.ProcessPasswordReset))
		defer server.Close()

		req, err := http.NewRequest("POST", server.URL+"/users/password/reset", bytes.NewBuffer(badPayload))
		if err != nil {
			t.Errorf("Unexpected error in get request to %s", req.URL)
		}
		response, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Errorf("Expected no error, got %s", err)
		}
		defer response.Body.Close()
		if response.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status %d, but got %d", http.StatusBadRequest, response.StatusCode)
		}
		want := "body contains bad JSON\n"
		body, err := io.ReadAll(response.Body)
		if err != nil {
			t.Errorf("Unexpected error reading response body: %v", err)
		}
		if strings.Contains(string(body), want) {
			t.Errorf("Expected body %s, but got %s", want, string(body))
		}
	})
	t.Run("Process password reset Invalid email", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(app.ProcessPasswordReset))
		defer server.Close()

		req, err := http.NewRequest("POST", server.URL+"/users/password/reset", bytes.NewBuffer(emailOnlyBadPayload))
		if err != nil {
			t.Errorf("Unexpected error in get request to %s", req.URL)
		}
		response, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Errorf("Expected no error, got %s", err)
		}
		defer response.Body.Close()
		if response.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status %d, but got %d", http.StatusBadRequest, response.StatusCode)
		}
		want := "body contains bad JSON\n"
		body, err := io.ReadAll(response.Body)
		if err != nil {
			t.Errorf("Unexpected error reading response body: %v", err)
		}
		if strings.Contains(string(body), want) {
			t.Errorf("Expected body %s, but got %s", want, string(body))
		}
	})
}

func TestApp_AuthenticateIntegration(t *testing.T) {
	testCfg := data.TestPostgresConfig()
	testDB, err := data.Open(testCfg)
	defer testDB.Close()
	if err != nil {
		t.Errorf("Expected database to open, but got %s", err)
	}
	app := App{userModel: &models.UserModel{DB: testDB}}
	t.Run("Authenticate Happy path", func(t *testing.T) {
		user := models.User{
			Email:     "deleteme",
			Password:  "deleteme",
			CreatedAt: time.Now(),
		}
		// just dummy insert it without making another request to make testing more efficent the add handler has already
		// been tested
		err := app.userModel.Insert(&user)
		if err != nil {
			t.Errorf("Expected no error, got %s", err)
		}
		server := httptest.NewServer(http.HandlerFunc(app.Authenticate))
		defer server.Close()
		var emailPayload = []byte(`{"email": "deleteme", "password": "deleteme"}`)
		req, err := http.NewRequest("POST", server.URL+"/users/login)", bytes.NewBuffer(emailPayload))
		if err != nil {
			t.Errorf("Unexpected error in get request to %s", req.URL)
		}
		resp, err := http.DefaultClient.Do(req)
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status %d, but got %d", http.StatusOK, resp.StatusCode)
		}

		var userReturned models.User
		err = json.NewDecoder(resp.Body).Decode(&userReturned)
		if err != nil {
			t.Errorf("Error unmarshaling JSON: %v", err)
		}
		userReturned.CreatedAt = user.CreatedAt
		if userReturned.Email != user.Email {
			t.Errorf("Expected user %v, but got %v", user, userReturned)
		}
		err = app.userModel.DeleteUser(user.Email)
		if err != nil {
			t.Errorf("Expected no error, got %s", err)
		}
	})
	t.Run("Authenticate Sad path", func(t *testing.T) {
		user := models.User{
			Email:     "deleteme",
			Password:  "deleteme",
			CreatedAt: time.Now(),
		}
		// insert user again as we deleted him in our initial test
		err := app.userModel.Insert(&user)
		if err != nil {
			t.Errorf("Expected no error, got %s", err)
		}
		server := httptest.NewServer(http.HandlerFunc(app.Authenticate))
		defer server.Close()
		var emailPayload = []byte(`{"email": "deleteme", "password": "wrongpassword"}`)
		req, err := http.NewRequest("POST", server.URL+"/users/login)", bytes.NewBuffer(emailPayload))
		if err != nil {
			t.Errorf("Unexpected error in get request to %s", req.URL)
		}
		resp, err := http.DefaultClient.Do(req)
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status %d, but got %d", http.StatusBadRequest, resp.StatusCode)
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Errorf("Unexpected error reading response body: %v", err)
		}
		want := "Invalid Credentials\n"
		if string(body) != want {
			t.Errorf("Expected body %s, but got %s", want, string(body))
		}
		err = app.userModel.DeleteUser(user.Email)
		if err != nil {
			t.Errorf("Expected no error, got %s", err)
		}
	})
	t.Run("Authenticate Invalid payload", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(app.Authenticate))
		defer server.Close()

		req, err := http.NewRequest("POST", server.URL+"/users/login", bytes.NewBuffer(badPayload))
		if err != nil {
			t.Errorf("Unexpected error in get request to %s", req.URL)
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status %d, but got %d", http.StatusBadRequest, resp.StatusCode)
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Errorf("Unexpected error reading response body: %v", err)
		}
		want := "body contains"
		if !strings.Contains(string(body), "body contains") {
			t.Errorf("Expected body %s, but got %s", want, string(body))
		}
	})
	t.Run("Authenticate Invalid email", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(app.Authenticate))
		defer server.Close()

		req, err := http.NewRequest("POST", server.URL+"/users/login", bytes.NewBuffer(badEmailPayload))
		if err != nil {
			t.Errorf("Unexpected error in get request to %s", req.URL)
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Errorf("Unexpected error reading response body: %v", err)
		}
		want := "User password must be 4 characters long and email must be 5 characters long\n"
		if string(body) != want {
			t.Errorf("Expected body %s, but got %s", want, string(body))
		}
	})

}

func TestApp_LogoutIntegration(t *testing.T) {
	testCfg := data.TestPostgresConfig()
	testDB, err := data.Open(testCfg)
	defer testDB.Close()
	if err != nil {
		t.Errorf("Expected database to open, but got %s", err)
	}
	app := App{userModel: &models.UserModel{DB: testDB}}

	t.Run("Logout happy path", func(t *testing.T) {
		// insert manually as we need to do for both tests
		user := models.User{
			Email:     "deleteme",
			Password:  "deleteme",
			CreatedAt: time.Now(),
		}
		err = app.userModel.Insert(&user)

		if err != nil {
			t.Errorf("Expected no error, got %s", err)
		}
		server := httptest.NewServer(http.HandlerFunc(app.Authenticate))
		defer server.Close()
		var emailPayload = []byte(`{"email": "deleteme", "password": "deleteme"}`)
		req, err := http.NewRequest("POST", server.URL+"/users/login)", bytes.NewBuffer(emailPayload))
		if err != nil {
			t.Errorf("Unexpected error in get request to %s", req.URL)
		}
		resp, err := http.DefaultClient.Do(req)
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status %d, but got %d", http.StatusOK, resp.StatusCode)
		}

		var userReturned models.User
		err = json.NewDecoder(resp.Body).Decode(&userReturned)
		if err != nil {
			t.Errorf("Error unmarshaling JSON: %v", err)
		}
		userReturned.CreatedAt = user.CreatedAt
		if userReturned.Email != user.Email {
			t.Errorf("Expected user %v, but got %v", user, userReturned)
		}
		cookie := resp.Cookies()[0]
		server = httptest.NewServer(http.HandlerFunc(app.SignOut))
		defer server.Close()
		req, err = http.NewRequest("POST", server.URL+"/users/logout)", nil)
		if err != nil {
			t.Errorf("Unexpected error in get request to %s", req.URL)
		}
		req.AddCookie(cookie)
		resp, err = http.DefaultClient.Do(req)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status %d, but got %d", http.StatusOK, resp.StatusCode)
		}
		want := "logged out succesfully"
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Errorf("Unexpected error reading response body: %v", err)
		}
		if strings.Contains(string(body), want) {
			t.Errorf("Expected body %s, but got %s", want, string(body))
		}
		err = app.userModel.DeleteUser(user.Email)
		if err != nil {
			t.Errorf("Expected no error, got %s", err)
		}
	})
	t.Run("Logout Sad path", func(t *testing.T) {
		// insert manually as we need to do for both tests
		user := models.User{
			Email:     "deleteme",
			Password:  "deleteme",
			CreatedAt: time.Now(),
		}
		err = app.userModel.Insert(&user)

		if err != nil {
			t.Errorf("Expected no error, got %s", err)
		}
		server := httptest.NewServer(http.HandlerFunc(app.Authenticate))
		defer server.Close()
		var emailPayload = []byte(`{"email": "deleteme", "password": "deleteme"}`)
		req, err := http.NewRequest("POST", server.URL+"/users/login)", bytes.NewBuffer(emailPayload))
		if err != nil {
			t.Errorf("Unexpected error in get request to %s", req.URL)
		}
		resp, err := http.DefaultClient.Do(req)
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status %d, but got %d", http.StatusOK, resp.StatusCode)
		}

		var userReturned models.User
		err = json.NewDecoder(resp.Body).Decode(&userReturned)
		if err != nil {
			t.Errorf("Error unmarshaling JSON: %v", err)
		}
		userReturned.CreatedAt = user.CreatedAt
		if userReturned.Email != user.Email {
			t.Errorf("Expected user %v, but got %v", user, userReturned)
		}

		server = httptest.NewServer(http.HandlerFunc(app.SignOut))
		defer server.Close()
		req, err = http.NewRequest("POST", server.URL+"/users/logout)", nil)
		if err != nil {
			t.Errorf("Unexpected error in get request to %s", req.URL)
		}
		resp, err = http.DefaultClient.Do(req)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("Expected status %d, but got %d", http.StatusUnauthorized, resp.StatusCode)
		}
		want := "unauthorized"
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Errorf("Unexpected error reading response body: %v", err)
		}
		if strings.Contains(string(body), want) {
			t.Errorf("Expected body %s, but got %s", want, string(body))
		}
		err = app.userModel.DeleteUser(user.Email)
		if err != nil {
			t.Errorf("Expected no error, got %s", err)
		}
	})
}
