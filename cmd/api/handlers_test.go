package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestApp_HandleHome(t *testing.T) {
	app := App{}

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Errorf("Unexpected error in get request to /")
	}

	rr := httptest.NewRecorder()

	app.HandleHome(rr, req)

	want := http.StatusOK
	got := rr.Code

	if got != want {
		t.Errorf("Got %d, want %d", got, want)
	}

	expected := "Welcome to the user service!"

	if rr.Body.String() != expected {
		t.Errorf("Got %s, wanted %s", rr.Body.String(), expected)
	}

}
