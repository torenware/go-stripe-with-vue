package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
)

func (app *application) routes() http.Handler {
	mux := chi.NewRouter()
	mux.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "PUT", "POST", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Allow", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	mux.Post("/api/payment-intent", app.GetPaymentIntent)
	mux.Post("/api/create-customer-and-subscribe-to-plan", app.ProcessSubscription)

	// Auth
	mux.Post("/api/authenticate", app.CreateAuthToken)
	mux.Post("/api/is-authenticated", app.CheckAuthentication)

	// To apply an auth middleware on a group of routes, we use the Router
	// method to create a sub-router

	return mux.Route("/api/auth", func(mux chi.Router) {
		mux.Use(app.AuthHandler)

		mux.Get("/vterm-success-handler", app.VTermSuccessHandler)
	})
}
