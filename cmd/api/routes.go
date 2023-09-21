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

	cookieMiddleware := app.RequireCookieMiddleware

	r.Group(func(r chi.Router) {
		r.Use(cookieMiddleware)
		r.Get("/", app.HandleHome)
		r.Get("/users", app.getUserByEmail)
	})

	r.Post("/users", app.CreateUser)
	r.Post("/users/login", app.Authenticate)
	r.Patch("/users", app.updateUserPassword)
	r.Post("/users/password/reset", app.ProcessPasswordReset)
	return r
}
