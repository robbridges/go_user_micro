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

func (app *App) enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// We can't guarantee the Access control header will be in every request, so add it as vary
		w.Header().Add("Vary", "Origin")
		w.Header().Add("Vary", "Access-Control-Request-Method")

		origin := r.Header.Get("Origin")

		if origin != "" {
			allowed := false
			for i := range app.Config.cors.trustedOrigins {
				if origin == app.Config.cors.trustedOrigins[i] {
					// Allow the origin.
					w.Header().Set("Access-Control-Allow-Origin", origin)
					allowed = true
					break
				}
			}

			if !allowed {
				// Origin is not in the trusted list, reject the request with a 403 Forbidden status.
				w.Header().Set("Access-Control-Allow-Origin", "https://example.com") // Set the disallowed origin.
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			// Handle preflight OPTIONS requests.
			if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
				w.Header().Set("Access-Control-Allow-Methods", "OPTIONS, PUT, PATCH, DELETE")
				w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
				w.WriteHeader(http.StatusOK)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}
