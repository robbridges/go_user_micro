package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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

}
