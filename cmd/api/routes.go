package main

import (
	"github.com/go-chi/chi/v5"
	"net/http"
)

func (app *App) SetRoutes() http.Handler {
	r := chi.NewRouter()

	//middleware
	r.Use(app.recoverPanic)

	r.Get("/", app.HandleHome)
	r.Post("/users", app.CreateUser)
	r.Patch("/users", app.updateUserPassword)
	r.Get("/users", app.getUserByEmail)

	return r
}
