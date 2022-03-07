package main

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/torenware/go-stripe/internal/models"
	"golang.org/x/crypto/bcrypt"
)

// writeJSON writes aribtrary data out as JSON
func (app *application) writeJSON(w http.ResponseWriter, status int, data interface{}, headers ...http.Header) error {
	out, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}

	if len(headers) > 0 {
		for k, v := range headers[0] {
			w.Header()[k] = v
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(out)

	return nil
}

// readJSON reads json from request body into data. We only accept a single json value in the body
func (app *application) readJSON(w http.ResponseWriter, r *http.Request, data interface{}) error {
	maxBytes := 1048576 // max one megabyte in request body
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	dec := json.NewDecoder(r.Body)
	err := dec.Decode(data)
	if err != nil {
		return err
	}

	// we only allow one entry in the json file
	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("body must only have a single JSON value")
	}

	return nil
}

// badRequest sends a JSON response with status http.StatusBadRequest, describing the error
func (app *application) badRequest(w http.ResponseWriter, r *http.Request, err error) error {
	var payload struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
	}

	payload.Error = true
	payload.Message = err.Error()

	out, err := json.MarshalIndent(payload, "", "\t")
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	w.Write(out)
	return nil
}

func (app *application) invalidCredentials(w http.ResponseWriter) error {
	var payload struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
	}
	payload.Error = true
	payload.Message = "invalid authentication credentials"

	err := app.writeJSON(w, http.StatusUnauthorized, &payload)
	return err
}

func (app *application) passwordsMatch(hash, offeredPW string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(offeredPW))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (app *application) getAuthenticatedUser(r *http.Request) (*models.User, error) {
	authHdr := r.Header.Get("Authorization")
	prefixLen := len("Bearer ")

	if len(authHdr) > 0 && authHdr[:prefixLen] == "Bearer " {
		token := authHdr[prefixLen:]
		user, err := app.DB.GetUserFromToken(token, AuthTokenTTL)
		if err != nil {
			if err.Error() != "token expired" {
				app.errorLog.Println(err)
				return nil, err
			}
			return nil, nil
		}
		return user, nil
	}
	return nil, errors.New("must supply auth header")
}
