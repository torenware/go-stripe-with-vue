package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (app *application) routes() http.Handler {
	mux := chi.NewMux()
	mux.Use(SessionLoad)

	mux.Get("/", app.HomePage)
	mux.Post("/payment-succeeded", app.PaymentSucceeded)
	mux.Get("/receipt", app.DisplayReceipt)

	mux.Get("/widget/{id}", app.BuyOneItem)
	mux.Get("/test-widget", app.TestGetWidget)

	mux.Get("/plans/bronze", app.BronzePlan)
	mux.Get("/receipt/bronze", app.ReceiptBronze)

	// Authentication
	mux.Get("/login", app.LoginPage)
	mux.Get("/logout", app.Logout)
	mux.Post("/process-login", app.ProcessLogin)

	mux.Route("/admin", func(mux chi.Router) {
		mux.Use(app.AuthHandler)
		mux.Get("/virtual-terminal", app.VirtualTerminal)
		mux.Post("/vterm-payment-succeeded", app.VTPaymentSucceeded)
	})

	fileServer := http.FileServer(http.Dir("./static/"))
	mux.Handle("/static/*", http.StripPrefix("/static", fileServer))

	return mux
}
