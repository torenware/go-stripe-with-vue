package main

import "net/http"

func (app *application) AuthHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, _ := app.getAuthenticatedUser(r)
		if user == nil {
			app.invalidCredentials(w)
			return
		}
		next.ServeHTTP(w, r)
		return
	})
}
