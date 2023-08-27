package main

import (
	"github.com/go-chi/chi/v5"
	"net/http"
)

func (app *App) SetRoutes() http.Handler {
	r := chi.NewRouter()

	r.Get("/", app.HandleHome)
	r.Post("/users", app.CreateUser)
	// Define more routes here

	return r
}
