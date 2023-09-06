package main

import (
	"fmt"
	"net/http"
)

func (app *App) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// if there was a panic, close the connection after response sent
				w.Header().Set("Connection", "close")

				http.Error(w, fmt.Sprintf("Panic recovered: %s", err), http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}
