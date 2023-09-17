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

func TestEnableCORS(t *testing.T) {
	// Create a new instance of your App struct with the desired CORS configuration.
	app := App{
		Config: Config{
			cors: struct {
				trustedOrigins []string
			}{
				trustedOrigins: []string{"http://localhost:3000"},
			},
		},
	}

	// Create a Chi router and apply the enableCORS middleware.
	r := chi.NewRouter()
	r.Use(app.enableCORS)

	// Define a sample route and handler.
	r.Route("/some-resource", func(r chi.Router) {
		r.MethodFunc(http.MethodOptions, "/", func(w http.ResponseWriter, r *http.Request) {
			// Your sample handler logic here for OPTIONS requests.
			w.WriteHeader(http.StatusOK)
		})
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			// Your sample handler logic here for GET requests.
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte("Response from /some-resource"))
			if err != nil {
				t.Errorf("Error writing response: %s", err)
			}
		})
	})

	t.Run("HappyPath", func(t *testing.T) {
		// Create an HTTP request to simulate a cross-origin request from an allowed origin.
		req := httptest.NewRequest(http.MethodOptions, "http://localhost:8080/some-resource", nil)
		req.Header.Set("Origin", "http://localhost:3000") // Set the trusted origin.

		// Create an HTTP response recorder to capture the response.
		rec := httptest.NewRecorder()

		// Serve the request through the Chi router.
		r.ServeHTTP(rec, req)

		// Verify the response. It should not have a CORS error status code.
		if rec.Code != http.StatusOK { // Adjust this based on your actual expected status code.
			t.Errorf("Expected status code %d, got %d", http.StatusOK, rec.Code)
		}
	})

	t.Run("SadPath", func(t *testing.T) {
		// Create an HTTP request to simulate a cross-origin request from a disallowed origin.
		req := httptest.NewRequest(http.MethodOptions, "http://localhost:8080/some-resource", nil)
		req.Header.Set("Origin", "https://example.com") // Set an origin not in the trusted list.

		// Create an HTTP response recorder to capture the response.
		rec := httptest.NewRecorder()

		// Serve the request through the Chi router.
		r.ServeHTTP(rec, req)

		// Verify the response. It should have a CORS error status code (403 Forbidden).
		if rec.Code != http.StatusForbidden { // Expect a 403 Forbidden status code.
			t.Errorf("Expected CORS error status code %d, got %d", http.StatusForbidden, rec.Code)
		}

		// Verify the CORS headers in the response.
		expectedHeaders := map[string]string{
			"Access-Control-Allow-Origin": "https://example.com", // Should match the request origin.
			// Include other expected CORS headers for the sad path as needed.
		}
		for key, value := range expectedHeaders {
			if rec.Header().Get(key) != value {
				t.Errorf("Expected header %s: %s, got %s", key, value, rec.Header().Get(key))
			}
		}
	})
}
