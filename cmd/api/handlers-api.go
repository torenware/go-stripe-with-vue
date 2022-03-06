package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/torenware/go-stripe/internal/cards"
	"github.com/torenware/go-stripe/internal/models"
)

const (
	AuthTokenTTL = 25 * time.Hour
)

type stripePayload struct {
	Currency      string `json:"currency"`
	Amount        int    `json:"amount"`
	PlanID        string `json:"plan"`
	PaymentMethod string `json:"payment_method"`
	Email         string `json:"email"`
	CardBrand     string `json:"card_brand"`
	ExpiryMonth   int    `json:"exp_month"`
	ExpiryYear    int    `json:"exp_year"`
	LastFour      string `json:"last_four"`
	ProductID     string `json:"product_id"`
	FirstName     string `json:"first_name"`
	LastName      string `json:"last_name"`
}

type jsonResponse struct {
	OK      bool   `json:"ok"`
	Message string `json:"message,omitempty"`
	Content string `json:"content,omitempty"`
	ID      int    `json:"id,omitempty"`
}

func (app *application) GetPaymentIntent(w http.ResponseWriter, r *http.Request) {
	var payload stripePayload

	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		// stub this until we implement a more reasonable
		// error handling strategy
		app.errorLog.Println(err)
		return
	}

	card := cards.Card{
		Secret:   app.config.stripe.secret,
		Key:      app.config.stripe.key,
		Currency: payload.Currency,
	}

	okay := true // optimism

	pi, msg, err := card.Charge(payload.Currency, payload.Amount)
	if err != nil {
		okay = false
	}

	if okay {
		out, err := json.MarshalIndent(pi, "", "  ")
		if err != nil {
			// again, replace later
			app.errorLog.Println(err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(out)
	} else {
		j := jsonResponse{
			OK:      false,
			Message: msg,
		}

		out, err := json.MarshalIndent(j, "", "   ")
		if err != nil {
			app.errorLog.Println(err)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write(out)
	}

}

func (app *application) ProcessSubscription(w http.ResponseWriter, r *http.Request) {
	var payload stripePayload

	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	debug, _ := json.MarshalIndent(payload, "", "    ")
	app.infoLog.Println(string(debug))

	card := cards.Card{
		Secret:   app.config.stripe.secret,
		Key:      app.config.stripe.key,
		Currency: payload.Currency,
	}

	ok := true
	//var subscription *stripe.Subscription
	txnMsg := "Transaction is successful"

	cust, msg, err := card.CreateCustomer(payload.PaymentMethod, payload.Email)
	if err != nil {
		app.errorLog.Println(msg, err)
		ok = false
		txnMsg = msg
	}
	if ok {
		subscription, err := card.SubscribeCustomer(cust, payload.PlanID, payload.Email, payload.LastFour, "")
		if err != nil {
			app.errorLog.Println(msg, err)
			ok = false
			txnMsg = "Subscription failed"
		}
		app.infoLog.Println("Subscribed as:", subscription.ID)
	}

	if ok {
		// save to DB...
		sp := payload
		cust_id, err := app.SaveCustomer(sp.FirstName, sp.LastName, sp.Email)
		if err != nil {
			app.errorLog.Println(err)
			return
		}
		txn := models.Transaction{
			Amount:              sp.Amount,
			Currency:            sp.Currency,
			PaymentMethod:       sp.PaymentMethod,
			LastFour:            sp.LastFour,
			ExpiryMonth:         sp.ExpiryMonth,
			ExpiryYear:          sp.ExpiryYear,
			TransactionStatusID: 2,
			CreatedAt:           time.Now(),
			UpdatedAt:           time.Now(),
		}
		txnID, err := app.SaveTxn(txn)
		if err != nil {
			app.errorLog.Println(err)
			return
		}
		pID, err := strconv.Atoi(sp.ProductID)
		if err != nil {
			app.errorLog.Println(err)
			return
		}
		order := models.Order{
			WidgetID:      pID,
			TransactionID: txnID,
			CustomerID:    cust_id,
			StatusID:      1,
			Quantity:      1,
			Amount:        sp.Amount,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}
		_, err = app.SaveOrder(order)
		if err != nil {
			app.errorLog.Println(err)
			return
		}
	}

	// stub
	j := jsonResponse{
		OK:      ok,
		Message: txnMsg,
	}
	out, err := json.MarshalIndent(j, "", "  ")
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(out)

}

// Rather than share the wrapper routines, we choose to copy the code. At least in
// theory, they could diverge between their backend and frontend versions. In any
// case, it's a PITA to do so, since we have two different "main" packages so we
// can create two different apps. So here we go:

// Helper routines for DB writes

func (app *application) SaveCustomer(firstName, lastName, email string) (int, error) {
	customer := models.Customer{
		FirstName: firstName,
		LastName:  lastName,
		Email:     email,
	}
	id, err := app.DB.InsertCustomer(customer)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (app *application) SaveOrder(order models.Order) (int, error) {
	id, err := app.DB.InsertOrder(order)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (app *application) SaveTxn(txn models.Transaction) (int, error) {
	id, err := app.DB.InsertTransaction(txn)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// Authentication

func (app *application) CreateAuthToken(w http.ResponseWriter, r *http.Request) {
	var userInput struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJSON(w, r, &userInput)
	if err != nil {
		app.badRequest(w, r, err)
		return
	}

	// See if we have such a user
	user, err := app.DB.GetUserByEmail(userInput.Email)
	if err != nil {
		app.invalidCredentials(w)
		return
	}
	matches, err := app.passwordsMatch(user.Password, userInput.Password)
	if err != nil {
		// Exceptional case
		app.errorLog.Println(err)
		return
	}
	if !matches {
		app.invalidCredentials(w)
		return
	}
	// Now generate our token
	token, err := models.GenerateToken(user.ID, AuthTokenTTL, models.ScopeAuthentication)
	if err != nil {
		app.errorLog.Println(err)
		app.badRequest(w, r, err)
		return
	}

	err = app.DB.InsertToken(token, user)
	if err != nil {
		app.errorLog.Println(err)
		app.badRequest(w, r, err)
	}

	var payload struct {
		Error   bool          `json:"error"`
		Message string        `json:"message"`
		Token   *models.Token `json:"authentication_token"`
	}

	payload.Error = false
	payload.Message = fmt.Sprintf("token for %s created", userInput.Email)
	payload.Token = token

	err = app.writeJSON(w, http.StatusOK, payload)
	if err != nil {
		app.errorLog.Println(err)
	}
}

func (app *application) CheckAuthentication(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Error      bool   `json:"error"`
		Message    string `json:"message"`
		Authorized bool   `json:"authorized"`
	}

	authHdr := r.Header.Get("Authorization")
	prefixLen := len("Bearer ")

	if authHdr[:prefixLen] == "Bearer " {
		token := authHdr[prefixLen:]
		email, err := app.DB.GetEmailFromToken(token, AuthTokenTTL)
		if err != nil {
			if err.Error() != "token expired" {
				app.errorLog.Println(err)
				payload.Error = true
			}
			_ = app.writeJSON(w, http.StatusUnauthorized, payload)
			return
		}
		payload.Authorized = true
		payload.Message = fmt.Sprintf("Authorized for %s", email)
		_ = app.writeJSON(w, http.StatusOK, payload)
		return
	}
	payload.Message = "no token supplied"
	_ = app.writeJSON(w, http.StatusUnauthorized, payload)

}
