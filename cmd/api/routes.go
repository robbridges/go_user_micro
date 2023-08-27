package main

import (
	"github.com/go-chi/chi/v5"
	"net/http"
)

func (app *App) SetRoutes() http.Handler {
	r := chi.NewRouter()

	r.Get("/", app.HandleHome)
	r.Post("/users", app.CreateUser)
	r.Get("/users", app.getUserByEmail)
	// Define more routes here

	return r
}
