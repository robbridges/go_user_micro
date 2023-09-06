package main

import (
	"github.com/go-chi/chi/v5"
	"net/http"
	"net/http/httptest"
	"testing"
	"the_lonely_road/models"
)

func TestRecoverPanicMiddleware(t *testing.T) {
	app := App{userModel: &models.UserModelMock{DB: []*models.User{}}}
	// Create a new Chi router and apply the recoverPanic middleware
	r := chi.NewRouter()
	r.Use(app.recoverPanic)

	// Define a route that intentionally panics
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		panic("This is a forced panic!")
	})

	// Create an HTTP test request for the root path
	req := httptest.NewRequest("GET", "/", nil)

	// Create a recorder to capture the response
	rr := httptest.NewRecorder()

	// Serve the request using the router
	r.ServeHTTP(rr, req)

	// Check the response status code
	if rr.Code != http.StatusInternalServerError {
		t.Errorf("Expected status code %d, but got %d", http.StatusInternalServerError, rr.Code)
	}

	// Check the response body
	expectedBody := "Panic recovered: This is a forced panic!\n"
	if rr.Body.String() != expectedBody {
		t.Errorf("Expected response body '%s', but got '%s'", expectedBody, rr.Body.String())
	}
}
