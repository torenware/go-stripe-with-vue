package main

import (
	"io/fs"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func ServeVueAssets(mux *chi.Mux, fileSys fs.ReadFileFS, pathToAssets string) error {
	sub, err := fs.Sub(fileSys, pathToAssets)
	if err != nil {
		return err
	}
	assetServer := http.FileServer(http.FS(sub))
	mux.Handle("/assets/*", http.StripPrefix("/assets", assetServer))
	return nil
}

func (app *application) routes() http.Handler {
	mux := chi.NewMux()
	mux.Use(SessionLoad)
	mux.Use(app.logRequest)

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
	mux.Get("/forgot-password", app.ForgotPassword)
	mux.Get("/login-link-sent", app.PasswordLinkSent)
	mux.Get("/reset-password", app.ResetPassword)

	mux.Route("/admin", func(mux chi.Router) {
		mux.Use(app.AuthHandler)
		mux.Get("/virtual-terminal", app.VirtualTerminal)
		mux.Post("/vterm-payment-succeeded", app.VTPaymentSucceeded)

		mux.Get("/all-sales", app.AllSales)
		mux.Get("/all-subscriptions", app.AllSubscriptions)
		mux.Get("/order/{id}", app.GetSale)
		mux.Get("/subscription/{id}", app.GetSubscription)

		mux.Get("/all-users", app.AllUsers)
		mux.Get("/user/{id:[0-9]+}", app.ShowUser)
		mux.Get("/user/{id:[0-9]+}/edit", app.EditUser)
		mux.Get("/user/new", app.NewUserForm)
	})

	fileServer := http.FileServer(http.Dir("./static/"))
	mux.Handle("/static/*", http.StripPrefix("/static", fileServer))

	err := ServeVueAssets(mux, dist, "dist/assets")
	if err != nil {
		app.errorLog.Println(err)
	}
	return mux
}
