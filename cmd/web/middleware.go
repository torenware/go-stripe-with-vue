package main

import (
	"net/http"
)

// SessionLoader implements the SCS middleware.
func SessionLoad(next http.Handler) http.Handler {
	return session.LoadAndSave(next)
}
