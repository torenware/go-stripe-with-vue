package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

func (app *application) GuardedFileServer(stripPrefix string, serveDir fs.FS) http.Handler {

	handler := func(w http.ResponseWriter, r *http.Request) {
		var err error
		var fileServer http.Handler
		fileFS := serveDir
		isEmbedded := false

		if _, ok := serveDir.(embed.FS); ok {
			fileFS, err = fs.Sub(serveDir, app.vueConfig.AssetsPath)
			if err != nil {
				log.Println("could not sub the FS", err)
				http.NotFound(w, r)
				return
			}
			isEmbedded = true
		}
		prefixLen := len(stripPrefix)
		rest := r.URL.Path[prefixLen:]
		// log.Println("via our file server:", rest)
		parts := strings.Split(rest, "/")
		// We want to prevent dot files from getting served.
		if parts[len(parts)-1][:1] == "." {
			//force a relative link.
			log.Printf("Found dotfile or dir %s", parts[0])
			http.NotFound(w, r)
			return
		}
		if isEmbedded {
			fileServer = http.FileServer(http.FS(fileFS))
		} else {
			fileServer = http.StripPrefix(stripPrefix, http.FileServer(http.FS(fileFS)))
		}
		fileServer.ServeHTTP(w, r)
	}

	return http.HandlerFunc(handler)
}

func (app *application) ServeVueAssets(mux *chi.Mux, prefix, stripPrefix string, serveDir fs.FS) error {
	assetServer := app.GuardedFileServer(stripPrefix, serveDir)
	mux.Handle(prefix + "*", assetServer)
	return nil
}


func (app *application) routes() http.Handler {
	mux := chi.NewMux()
	mux.Use(SessionLoad)
	mux.Use(app.RequestLoggerMiddleware)

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

	err := app.ServeVueAssets(mux, app.vueConfig.URLPrefix, "/", app.vueConfig.FS)
	if err != nil {
		app.errorLog.Println(err)
	}
	return mux
}
