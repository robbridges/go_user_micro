package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type TestStruct struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

func TestWriteJSON(t *testing.T) {
	app := &App{} // Replace with your actual App initialization

	t.Run("ValidData", func(t *testing.T) {
		// Prepare valid JSON data
		data := TestStruct{Name: "test", Value: 42}

		// Create an HTTP response recorder
		w := httptest.NewRecorder()

		// Call the writeJSON function
		err := app.writeJSON(w, http.StatusOK, data)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		// Verify the response status code
		if w.Code != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
		}

		// Verify the Content-Type header
		contentType := w.Header().Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("Expected Content-Type 'application/json', got '%s'", contentType)
		}

		// Verify the response body
		var response TestStruct
		err = json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Errorf("Failed to unmarshal response body: %v", err)
		}

		if response.Name != data.Name || response.Value != data.Value {
			t.Errorf("Expected response %+v, got %+v", data, response)
		}
	})

	t.Run("InvalidData", func(t *testing.T) {
		// Prepare invalid JSON data (a map with an unsupported type)
		invalidData := map[string]interface{}{
			"name":  "test",
			"value": func() {}, // Unsupported type
		}

		// Create an HTTP response recorder
		w := httptest.NewRecorder()

		// Call the writeJSON function
		err := app.writeJSON(w, http.StatusOK, invalidData)

		if err == nil {
			t.Errorf("Expected an error, got nil")
		}

		// Verify the response status code (should still be set to 200)
		if w.Code != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
		}
	})
}

func TestReadJSON(t *testing.T) {
	app := &App{} // Replace with your actual App initialization

	t.Run("ValidJSON", func(t *testing.T) {
		// Prepare a valid JSON request body
		validJSON := `{"name": "test", "value": 42}`
		req := httptest.NewRequest("POST", "/test", strings.NewReader(validJSON))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		var dst TestStruct
		err := app.readJSON(w, req, &dst)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if dst.Name != "test" || dst.Value != 42 {
			t.Errorf("Expected dst to be {\"name\": \"test\", \"value\": 42}, got %+v", dst)
		}
	})

	t.Run("MaxBytesExceeded", func(t *testing.T) {
		// Prepare a request with a JSON body larger than maxBytes
		maxBytesExceededJSON := `{"data": "` + strings.Repeat("x", 1_048_577) + `"}`
		req := httptest.NewRequest("POST", "/test", strings.NewReader(maxBytesExceededJSON))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		var dst TestStruct
		err := app.readJSON(w, req, &dst)

		want := "got body must not be larger than 1048576 bytes"
		if err == nil || strings.Contains(err.Error(), want) {
			t.Errorf("Expected error %q, got %v", want, err)
		}
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		// Prepare an invalid JSON request body
		invalidJSON := `{"name": "test", "value": "not_an_int"}`
		req := httptest.NewRequest("POST", "/test", strings.NewReader(invalidJSON))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		var dst TestStruct
		err := app.readJSON(w, req, &dst)

		if err == nil {
			t.Errorf("Expected error, got %v", err)
		}
	})

	t.Run("UnknownField", func(t *testing.T) {
		// Prepare a request with an unknown field
		unknownFieldJSON := `{"name": "test", "value": 42, "extra": "field"}`
		req := httptest.NewRequest("POST", "/test", strings.NewReader(unknownFieldJSON))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		var dst TestStruct
		err := app.readJSON(w, req, &dst)

		if !strings.Contains(err.Error(), "body contains unknown key") {
			t.Errorf("Expected error message containing 'body contains unknown key', got %v", err)
		}
	})

	t.Run("EmptyBody", func(t *testing.T) {
		// Prepare a request with an empty body
		req := httptest.NewRequest("POST", "/test", nil)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		var dst TestStruct
		err := app.readJSON(w, req, &dst)

		if err == nil || err.Error() != "body must not be empty" {
			t.Errorf("Expected 'body must not be empty' error, got %v", err)
		}
	})

	t.Run("MultipleValues", func(t *testing.T) {
		// Prepare a request with multiple JSON values in the body
		multipleValuesJSON := `{"name": "test", "value": 42}{"name": "test2", "value": 43}`
		req := httptest.NewRequest("POST", "/test", strings.NewReader(multipleValuesJSON))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		var dst TestStruct
		err := app.readJSON(w, req, &dst)

		if err == nil || err.Error() != "body must only contain a single JSON value" {
			t.Errorf("Expected 'body must only contain a single JSON value' error, got %v", err)
		}
	})
}
