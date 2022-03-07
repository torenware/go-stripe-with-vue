package main

import (
	"net/http"
)

// SessionLoader implements the SCS middleware.
func SessionLoad(next http.Handler) http.Handler {
	return session.LoadAndSave(next)
}

func (app *application) AuthHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !session.Exists(r.Context(), "userID") {
			http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		}
		next.ServeHTTP(w, r)
	})
}
