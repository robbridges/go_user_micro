package main

import (
	"github.com/go-chi/chi/v5"
	"net/http"
)

func (app *App) SetRoutes() http.Handler {
	r := chi.NewRouter()

	//middleware
	r.Use(app.recoverPanic)
	r.Use(app.enableCORS)

	r.Group(func(r chi.Router) {
		r.Use(app.RequireCookieMiddleware)
		r.Get("/", app.HandleHome)
		r.Get("/users", app.getUserByEmail)
	})

	r.Post("/users", app.CreateUser)
	r.Post("/users/login", app.Authenticate)
	r.Patch("/users", app.updateUserPassword)
	r.Post("/users/password/reset", app.ProcessPasswordReset)
	r.Post("/users/logout", app.SignOut)
	return r
}
