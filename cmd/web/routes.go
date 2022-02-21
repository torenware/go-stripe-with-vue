package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (app *application) routes() http.Handler {
	mux := chi.NewMux()

	mux.Get("/virtual-terminal", app.VirtualTerminal)

	return mux
}
