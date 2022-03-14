package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/stripe/stripe-go/v72"
	"net/http"
	"strconv"
	"time"

	"github.com/torenware/go-stripe/internal/cards"
	"github.com/torenware/go-stripe/internal/models"
	"github.com/torenware/go-stripe/internal/urlsigner"
	"golang.org/x/crypto/bcrypt"
)

const (
	AuthTokenTTL = 24 * time.Hour
)

type stripePayload struct {
	Currency      string `json:"currency"`
	Amount        int    `json:"amount"`
	PlanID        string `json:"plan"`
	PaymentMethod string `json:"payment_method"`
	PaymentIntent string `json:"payment_intent"`
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
		_, _ = w.Write(out)
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
		_, _ = w.Write(out)
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
	var subscription *stripe.Subscription
	txnMsg := "Transaction is successful"

	cust, msg, err := card.CreateCustomer(payload.PaymentMethod, payload.Email)
	if err != nil {
		app.errorLog.Println(msg, err)
		ok = false
		txnMsg = msg
	}
	if ok {
		subscription, err = card.SubscribeCustomer(cust, payload.PlanID, payload.Email, payload.LastFour, "")
		if err != nil {
			app.errorLog.Println(msg, err)
			ok = false
			txnMsg = "Subscription failed"
		}
	}

	// Update the record to save a few fields. We use the PaymentsIntent field to hold the sub ID,
	// and save the pm as well.
	if ok {
		// save to DB...
		sp := payload
		custID, err := app.SaveCustomer(sp.FirstName, sp.LastName, sp.Email)
		if err != nil {
			app.errorLog.Println(err)
			return
		}
		txn := models.Transaction{
			Amount:              sp.Amount,
			Currency:            sp.Currency,
			PaymentMethod:       sp.PaymentMethod,
			PaymentIntent:       subscription.ID, // we reuse this field. Not my idea :-)
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
			CustomerID:    custID,
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
	_, _ = w.Write(out)

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

func (app *application) PasswordLink(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Email string `json:"email"`
	}

	err := app.readJSON(w, r, &payload)
	if err != nil {
		_ = app.badRequest(w, r, err)
		return
	}

	// See if we have such a user
	user, err := app.DB.GetUserByEmail(payload.Email)
	if err != nil {
		_ = app.invalidCredentials(w)
		return
	}

	// We say we are sending even if we are not...
	var output struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
	}

	var data struct {
		Link string
	}

	link := fmt.Sprintf("%s/reset-password?email=%s", app.config.frontend, payload.Email)
	sign := urlsigner.Signer{
		Secret: []byte(app.config.secretkey),
	}

	signedLink := sign.GenerateTokenFromString(link)
	data.Link = signedLink

	if user.ID != 0 {
		app.infoLog.Println("Email would be sent here")
		// send mail
		err = app.SendMail("info@widgets.com", user.Email, "Password Reset Request", "password-reset", data)
		if err != nil {
			app.errorLog.Println(err)
			_ = app.badRequest(w, r, err)
			return
		}
		output.Error = false
		output.Message = "We've sent you an email with a link"
	} else {
		output.Error = true
		output.Message = "Sorry! We cannot find you in our system."
	}

	_ = app.writeJSON(w, http.StatusOK, output)

}

func (app *application) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Email     string `json:"email"`
		EmailHash string `json:"email_hash"`
		Password  string `json:"password"`
	}

	err := app.readJSON(w, r, &payload)
	if err != nil {
		_ = app.badRequest(w, r, err)
		return
	}

	// Does it look legit?
	sign := urlsigner.Signer{
		Secret: []byte(app.config.secretkey),
	}
	err = sign.ConfirmHashForString(payload.EmailHash, payload.Email)
	if err != nil {
		app.errorLog.Println("validation failed:", err)
		// but we continue...
	} else {
		app.infoLog.Println("validation worked!")
	}
	user, err := app.DB.GetUserByEmail(payload.Email)
	if err != nil {
		_ = app.badRequest(w, r, err)
		return
	}

	newHash, err := bcrypt.GenerateFromPassword([]byte(payload.Password), 12)
	if err != nil {
		_ = app.badRequest(w, r, err)
		return
	}

	// see if it is valid
	err = bcrypt.CompareHashAndPassword(newHash, []byte(payload.Password))
	app.infoLog.Println("CHAP returned", err)

	newUser, _ := app.DB.GetUserByEmail(payload.Email)
	if newUser.Password != string(newHash) {
		app.infoLog.Println("pw correctly reset")
	} else {
		app.infoLog.Println("pw got munged!!")
	}

	err = app.DB.UpdatePasswordForUser(user, string(newHash))
	if err != nil {
		_ = app.badRequest(w, r, err)
		return
	}

	var output struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
	}

	output.Error = false
	output.Message = "Email has been changed"

	_ = app.writeJSON(w, http.StatusOK, output)

}

func (app *application) CreateAuthToken(w http.ResponseWriter, r *http.Request) {
	var userInput struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJSON(w, r, &userInput)
	if err != nil {
		_ = app.badRequest(w, r, err)
		return
	}

	// See if we have such a user
	user, err := app.DB.GetUserByEmail(userInput.Email)
	if err != nil {
		_ = app.invalidCredentials(w)
		return
	}
	matches, err := app.passwordsMatch(user.Password, userInput.Password)
	if err != nil {
		// Exceptional case
		app.errorLog.Println(err)
		return
	}
	if !matches {
		_ = app.invalidCredentials(w)
		return
	}
	// Now generate our token
	token, err := models.GenerateToken(user.ID, AuthTokenTTL, models.ScopeAuthentication)
	if err != nil {
		app.errorLog.Println(err)
		_ = app.badRequest(w, r, err)
		return
	}

	err = app.DB.InsertToken(token, user)
	if err != nil {
		app.errorLog.Println(err)
		_ = app.badRequest(w, r, err)
	}

	var payload struct {
		Error   bool          `json:"error"`
		Message string        `json:"message"`
		Token   *models.Token `json:"authentication_token"`
		UserID  int           `json:"user_id,omitempty"`
	}

	payload.Error = false
	payload.Message = fmt.Sprintf("token for %s created", userInput.Email)
	payload.Token = token
	payload.UserID = user.ID

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
		user, err := app.DB.GetUserFromToken(token, AuthTokenTTL)
		if err != nil {
			if err.Error() != "token expired" {
				app.errorLog.Println(err)
				payload.Error = true
			}
			_ = app.writeJSON(w, http.StatusUnauthorized, payload)
			return
		}
		payload.Authorized = true
		payload.Message = fmt.Sprintf("Authorized for %s", user.Email)
		_ = app.writeJSON(w, http.StatusOK, payload)
		return
	}
	payload.Message = "no token supplied"
	_ = app.writeJSON(w, http.StatusUnauthorized, payload)

}

func (app *application) CreateNewUser(w http.ResponseWriter, r *http.Request) {
	var userInput struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Email     string `json:"email"`
		Password  string `json:"password"`
	}

	err := app.readJSON(w, r, &userInput)
	test, _ := json.MarshalIndent(userInput, "", "  ")
	app.infoLog.Println(string(test))

	// Does the user already exist on this email?
	user, err := app.DB.GetUserByEmail(userInput.Email)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			_ = app.badRequest(w, r, err)
			return
		}
	}
	if user.ID != 0 {
		_ = app.badRequest(w, r, errors.New("email already in use"))
		return
	}

	// Might want to validate the rest of these...

	newHash, err := bcrypt.GenerateFromPassword([]byte(userInput.Password), 12)
	if err != nil {
		_ = app.badRequest(w, r, err)
		return
	}

	var u models.User
	u.FirstName = userInput.FirstName
	u.LastName = userInput.LastName
	u.Email = userInput.Email
	u.Password = string(newHash)

	uid, err := app.DB.InsertUser(u)
	if err != nil {
		_ = app.badRequest(w, r, err)
		return
	}

	var out struct {
		Error   string `json:"error"`
		Message string `json:"message"`
		UserID  int    `json:"user_id"`
	}

	out.Message = fmt.Sprintf("user created at id=%d", uid)
	out.UserID = uid

	_ = app.writeJSON(w, http.StatusCreated, out)
}

func (app *application) VTermSuccessHandler(w http.ResponseWriter, r *http.Request) {
	var txnData struct {
		PaymentAmount   int    `json:"payment_amount"`
		PaymentCurrency string `json:"payment_currency"`
		FirstName       string `json:"first_name"`
		LastName        string `json:"last_name"`
		PaymentIntent   string `json:"payment_intent"`
		PaymentMethod   string `json:"payment_method"`
		ExpiryMonth     int    `json:"expiry_month"`
		ExpiryYear      int    `json:"expiry_year"`
		LastFour        string `json:"last_four"`
		BankReturnCode  string `json:"bank_return_code"`
	}

	err := app.readJSON(w, r, &txnData)
	if err != nil {
		app.errorLog.Println("RJ", err)
		_ = app.badRequest(w, r, err)
		return
	}

	card := cards.Card{
		Secret: app.config.stripe.secret,
		Key:    app.config.stripe.key,
	}

	pi, err := card.RetrievePaymentIntent(txnData.PaymentIntent)
	if err != nil {
		app.errorLog.Println("RPI", err)
		_ = app.badRequest(w, r, err)
		return
	}

	pm, err := card.GetPaymentMethod(txnData.PaymentMethod)
	if err != nil {
		app.errorLog.Println("GPM", err)
		_ = app.badRequest(w, r, err)
		return
	}

	txnData.LastFour = pm.Card.Last4
	txnData.ExpiryMonth = int(pm.Card.ExpMonth)
	txnData.ExpiryYear = int(pm.Card.ExpYear)
	txnData.BankReturnCode = pi.Charges.Data[0].ID

	txn := models.Transaction{
		Amount:              txnData.PaymentAmount,
		Currency:            txnData.PaymentCurrency,
		LastFour:            txnData.LastFour,
		ExpiryMonth:         txnData.ExpiryMonth,
		ExpiryYear:          txnData.ExpiryYear,
		BankReturnCode:      txnData.BankReturnCode,
		PaymentIntent:       txnData.PaymentIntent,
		PaymentMethod:       txnData.PaymentMethod,
		TransactionStatusID: 2,
	}

	id, err := app.SaveTxn(txn)
	if err != nil {
		app.errorLog.Println("STX", err)
		_ = app.badRequest(w, r, err)
		return
	}
	txn.ID = id
	_ = app.writeJSON(w, http.StatusOK, txn)
}

func (app *application) ListSales(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		PageSize    int `json:"page_size"`
		CurrentPage int `json:"current_page"` // 1 based
	}
	err := app.readJSON(w, r, &payload)
	if err != nil {
		_ = app.badRequest(w, r, err)
		return
	}

	if payload.PageSize > 0 && payload.CurrentPage < 1 {
		err = errors.New("page must be 1 or greater")
		_ = app.badRequest(w, r, err)
		return
	}

	rows, lastPage, totalRows, err := app.DB.GetPaginatedSales(payload.PageSize, payload.CurrentPage)
	if err != nil {
		_ = app.badRequest(w, r, err)
		return
	}

	var out struct {
		Error       bool            `json:"error"`
		Rows        []*models.Order `json:"rows"`
		CurrentPage int             `json:"current_page"`
		LastPage    int             `json:"last_page"`
		TotalRows   int             `json:"total_rows"`
	}

	out.Rows = rows
	out.LastPage = lastPage
	out.TotalRows = totalRows
	out.CurrentPage = payload.CurrentPage

	_ = app.writeJSON(w, http.StatusOK, out)
}

func (app *application) ListSubscriptions(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		PageSize    int `json:"page_size"`
		CurrentPage int `json:"current_page"` // 1 based
	}
	err := app.readJSON(w, r, &payload)
	if err != nil {
		_ = app.badRequest(w, r, err)
		return
	}

	if payload.PageSize > 0 && payload.CurrentPage < 1 {
		err = errors.New("page must be 1 or greater")
		_ = app.badRequest(w, r, err)
		return
	}

	rows, lastPage, totalRows, err := app.DB.GetPaginatedSubscriptions(payload.PageSize, payload.CurrentPage)
	if err != nil {
		_ = app.badRequest(w, r, err)
		return
	}

	var out struct {
		Error       bool            `json:"error"`
		Rows        []*models.Order `json:"rows"`
		CurrentPage int             `json:"current_page"`
		LastPage    int             `json:"last_page"`
		TotalRows   int             `json:"total_rows"`
	}

	out.Rows = rows
	out.LastPage = lastPage
	out.TotalRows = totalRows
	out.CurrentPage = payload.CurrentPage

	_ = app.writeJSON(w, http.StatusOK, out)
}

func (app *application) ListUsers(w http.ResponseWriter, r *http.Request) {
	// Not going to bother with pagination.

	var output struct {
		Error   string         `json:"error"`
		Message string         `json:"message"`
		Users   []*models.User `json:"users"`
	}

	users, err := app.DB.GetAllUsers()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			var emptyRows []*models.User
			// not really an error
			output.Message = "no rows found"
			output.Users = emptyRows
			_ = app.writeJSON(w, http.StatusOK, output)
			return
		}
		_ = app.badRequest(w, r, err)
	}
	output.Users = users
	_ = app.writeJSON(w, http.StatusOK, output)
}

func (app *application) SingleSale(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		app.errorLog.Println("url param must be an integer")
		_ = app.badRequest(w, r, errors.New("url param must be an integer"))
		return
	}
	item, err := app.DB.GetSale(id)
	if err != nil {
		app.errorLog.Println(err.Error())
		_ = app.badRequest(w, r, err)
		return
	}
	_ = app.writeJSON(w, http.StatusOK, item)
}

func (app *application) SingleSubscription(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		app.errorLog.Println("url param must be an integer")
		_ = app.badRequest(w, r, errors.New("url param must be an integer"))
		return
	}
	item, err := app.DB.GetSubscription(id)
	if err != nil {
		app.errorLog.Println(err.Error())
		_ = app.badRequest(w, r, err)
		return
	}
	_ = app.writeJSON(w, http.StatusOK, item)
}

func (app *application) RefundCharge(w http.ResponseWriter, r *http.Request) {
	var chargeToRefund struct {
		ID            int    `json:"id"` // order_Id
		PaymentIntent string `json:"pi"`
		Amount        int    `json:"amount"` // 0 for full.
		Currency      string `json:"currency"`
	}
	err := app.readJSON(w, r, &chargeToRefund)
	if err != nil {
		_ = app.badRequest(w, r, err)
		return
	}

	// Let's validate the request.
	order, err := app.DB.GetSale(chargeToRefund.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			app.notFound(w, r)
			return
		}
		_ = app.badRequest(w, r, err)
		return
	}

	if chargeToRefund.Amount == 0 {
		chargeToRefund.Amount = order.Amount
	} else if chargeToRefund.Amount > order.Amount {
		// fraud
		app.errorLog.Println("FRAUD: overrefunding a charge")
		_ = app.badRequest(w, r, errors.New("rejected"))
		return
	}

	card := cards.Card{
		Secret:   app.config.stripe.secret,
		Key:      app.config.stripe.key,
		Currency: order.Transaction.Currency,
	}
	err = card.Refund(chargeToRefund.PaymentIntent, chargeToRefund.Amount)
	if err != nil {
		_ = app.badRequest(w, r, err)
		return
	}
	err = app.DB.SetOrderStatusID(order.ID, cards.STATUS_REFUNDED)
	if err != nil {
		_ = app.badRequest(w, r, err)
		return
	}

	var resp struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
	}

	resp.Error = false
	resp.Message = "Refund created"
	_ = app.writeJSON(w, http.StatusCreated, resp)
}

func (app *application) CancelSubscription(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		OrderID int `json:"id"`
	}
	err := app.readJSON(w, r, &payload)
	if err != nil {
		_ = app.badRequest(w, r, err)
		return
	}
	order, err := app.DB.GetSubscription(payload.OrderID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			app.notFound(w, r)
			return
		}
		_ = app.badRequest(w, r, err)
		return
	}

	card := cards.Card{
		Secret:   app.config.stripe.secret,
		Key:      app.config.stripe.key,
		Currency: order.Transaction.Currency,
	}
	// We stash the subID in the paymentIntent:
	err = card.CancelSubscription(order.Transaction.PaymentIntent)
	if err != nil {
		_ = app.badRequest(w, r, err)
		return
	}

	// update order status
	err = app.DB.SetOrderStatusID(order.ID, cards.STATUS_CANCELLED_SUB)
	if err != nil {
		_ = app.badRequest(w, r, err)
		return
	}

	var out struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
	}
	out.Message = "unsubscribe successful"

	_ = app.writeJSON(w, http.StatusOK, out)
}
