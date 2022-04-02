package main

import (
	"log"
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

func (app *application) logRequest(next http.Handler) http.Handler {
	app.infoLog.Println("invoked log handler")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.infoLog.Printf("%s - %s %s %s", r.RemoteAddr, r.Proto, r.Method, r.URL.RequestURI())

		next.ServeHTTP(w, r)
	})
}

// Fancy logger that knows about the status code:
// @see https://stackoverflow.com/questions/66528234/log-http-responsewriter-content

type ResponseWriterWrapper struct {
	w          *http.ResponseWriter
	statusCode *int
	headers    http.Header
	logger     *log.Logger
}

// RequestLoggerMiddleware is the middleware layer to log all the HTTP requests
func (app *application) RequestLoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rww := app.NewResponseWriterWrapper(w)
		w.Header()
		defer func() {
			rww.logger.Printf("%s - %s %s %s (%d)", r.RemoteAddr, r.Proto, r.Method, r.URL.RequestURI(), *rww.statusCode)
		}()
		next.ServeHTTP(rww, r)
	})
}

func (app *application) NewResponseWriterWrapper(w http.ResponseWriter) ResponseWriterWrapper {
	var statusCode int = 200
	return ResponseWriterWrapper{
		w:          &w,
		statusCode: &statusCode,
		headers:    w.Header(),
		logger:     app.infoLog,
	}
}

// Make sure we implement the interface for ResponseWriter
func (rww ResponseWriterWrapper) Write(buf []byte) (int, error) {
	return (*rww.w).Write(buf)
}

// Header function overwrites the http.ResponseWriter Header() function
func (rww ResponseWriterWrapper) Header() http.Header {
	return (*rww.w).Header()
}

// WriteHeader function overwrites the http.ResponseWriter WriteHeader() function
func (rww ResponseWriterWrapper) WriteHeader(statusCode int) {
	(*rww.statusCode) = statusCode
	(*rww.w).WriteHeader(statusCode)
}
