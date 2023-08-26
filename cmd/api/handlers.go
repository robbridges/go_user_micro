package main

import "net/http"

func (app *App) HandleHome(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Welcome to the user service!"))

}
