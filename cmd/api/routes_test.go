package main

import "testing"

func TestApp_SetRoutes(t *testing.T) {
	app := App{}

	router := app.SetRoutes()

	if router == nil {
		t.Errorf("Expected non nil pointer but got nil")
	}
}
