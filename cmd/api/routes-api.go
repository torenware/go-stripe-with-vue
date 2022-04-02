package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
)

func (app *application) routes() http.Handler {
	mux := chi.NewRouter()
	mux.Use(app.RequestLoggerMiddleware)
	mux.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "PUT", "POST", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Allow", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: false,
		MaxAge:           300,
		Debug:            false,
	}))

	mux.Post("/api/payment-intent", app.GetPaymentIntent)
	mux.Get("/api/sparams", app.StripeParams)
	mux.Post("/api/create-customer-and-subscribe-to-plan", app.ProcessSubscription)

	// Auth
	mux.Post("/api/authenticate", app.CreateAuthToken)
	mux.Post("/api/is-authenticated", app.CheckAuthentication)
	mux.Post("/api/password-link", app.PasswordLink)
	mux.Post("/api/reset-password", app.ResetPassword)

	// To apply an auth middleware on a group of routes, we use the Router
	// method to create a sub-router

	mux.Route("/api/auth", func(mux chi.Router) {
		mux.Use(app.AuthHandler)

		mux.Post("/vterm-success-handler", app.VTermSuccessHandler)
		mux.Post("/list-sales", app.ListSales)
		mux.Post("/list-subs", app.ListSubscriptions)
		mux.Post("/list-users", app.ListUsers)

		mux.Get("/sale/{id}", app.SingleSale)
		mux.Get("/subscription/{id}", app.SingleSubscription)

		mux.Post("/refund", app.RefundCharge)
		mux.Post("/cancel-subscription", app.CancelSubscription)

		mux.Post("/new-user", app.CreateNewUser)
		mux.Get("/user/{id}", app.SingleUser)
		mux.Post("/user/{id}", app.UpdateUser)
		mux.Delete("/user/{id}", app.DeleteUser)
	})

	return mux
}
