package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"the_lonely_road/data"
	"the_lonely_road/models"
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
		var user models.User
		err = json.NewDecoder(resp.Body).Decode(&user)
		if err != nil {
			t.Errorf("Error unmarshaling JSON: %v", err)
		}
		expectedEmail := "test@example.com"

		if user.Email != expectedEmail {
			t.Errorf("Expected email %s, but got %s", expectedEmail, user.Email)
		}
		app.userModel.DeleteUser(user.Email)
	})
	t.Run("Duplicate user", func(t *testing.T) {
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
		var user models.User
		err = json.NewDecoder(resp.Body).Decode(&user)
		if err != nil {
			t.Errorf("Error unmarshaling JSON: %v", err)
		}
		err = app.userModel.Insert(&user)
		if err == nil {
			t.Errorf("Expected error, but got nil")
		}
		resp, err = http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Errorf("Unexpected error reading response body: %v", err)
		}

		if string(body) != "duplicate email\n" {
			t.Errorf("Expected body %s, but got %s", "duplicate email\n", string(body))
		}
		app.userModel.DeleteUser(user.Email)
	})
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
		payload = []byte(`{"email": "admin@localhost"}`)
		req, err := http.NewRequest("GET", server.URL+"/users)", bytes.NewBuffer(payload))
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
		payload = []byte(`{"email": "adminx@localhost"}`)
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
}
