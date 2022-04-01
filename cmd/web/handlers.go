package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/torenware/go-stripe/internal/cards"
	"github.com/torenware/go-stripe/internal/models"
	"github.com/torenware/go-stripe/internal/urlsigner"
)

// The clientError helper sends a specific status code and corresponding description
// to the user. We'll use this later in the book to send responses like 400 "Bad
// Request" when there's a problem with the request that the user sent.
func (app *application) clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

//
func (app *application) setFlashAndGoHome(w http.ResponseWriter, r *http.Request, msg string, retCode int) {
	SetFlash(w, "flash", []byte(msg))
	http.Redirect(w, r, "/", retCode)
}

func (app *application) HomePage(w http.ResponseWriter, r *http.Request) {
	// This page has Vue support
	td := &templateData{}
	if app.vueglue != nil {
		td.VueGlue = app.vueglue
	}
	if err := app.renderTemplate(w, r, "home", td); err != nil {
		app.errorLog.Println(err)
		app.clientError(w, http.StatusBadRequest)
	}
}

func (app *application) VirtualTerminal(w http.ResponseWriter, r *http.Request) {
	// stub for now
	app.infoLog.Println("Hit VT endpoint")

	if err := app.renderTemplate(w, r, "terminal", nil, "stripejs", "stripe-form"); err != nil {
		app.errorLog.Println(err)
		app.clientError(w, http.StatusBadRequest)
	}
}

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

type TransactionData struct {
	ID              int
	FirstName       string
	LastName        string
	NameOnCard      string
	Email           string
	PaymentIntentID string
	PaymentMethodID string
	PaymentAmount   int
	PaymentCurrency string
	LastFour        string
	ExpiryMonth     int
	ExpiryYear      int
	BankReturnCode  string
}

func (app *application) GetTxnData(r *http.Request) (*TransactionData, error) {
	cardHolder := r.Form.Get("cardholder_name")
	email := r.Form.Get("email")
	firstName := r.Form.Get("first_name")
	lastName := r.Form.Get("last_name")
	paymentIntent := r.Form.Get("payment_intent")
	paymentMethod := r.Form.Get("payment_method")
	paymentCurrency := r.Form.Get("payment_currency")
	paymentAmount, err := strconv.Atoi(r.Form.Get("payment_amount"))
	if err != nil {
		app.errorLog.Println("payment amount is not an int")
		return nil, err
	}

	card := cards.Card{
		Secret: app.config.stripe.secret,
		Key:    app.config.stripe.key,
	}

	pi, err := card.RetrievePaymentIntent(paymentIntent)
	if err != nil {
		app.errorLog.Println(err)
		return nil, err
	}

	pm, err := card.GetPaymentMethod(paymentMethod)
	if err != nil {
		app.errorLog.Println(err)
		return nil, err
	}

	lastFour := pm.Card.Last4
	expiryMonth := pm.Card.ExpMonth
	expiryYear := pm.Card.ExpYear
	bankReturnCode := pi.Charges.Data[0].ID

	// worth doing validation here...

	txn := TransactionData{
		FirstName:       firstName,
		LastName:        lastName,
		NameOnCard:      cardHolder,
		Email:           email,
		PaymentIntentID: paymentIntent,
		PaymentMethodID: paymentMethod,
		PaymentAmount:   paymentAmount,
		PaymentCurrency: paymentCurrency,
		LastFour:        lastFour,
		ExpiryMonth:     int(expiryMonth),
		ExpiryYear:      int(expiryYear),
		BankReturnCode:  bankReturnCode,
	}

	return &txn, nil
}

func (app *application) VTPaymentSucceeded(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	txnPtr, err := app.GetTxnData(r)
	if err != nil {
		app.errorLog.Println(err)
		app.clientError(w, http.StatusBadRequest)
		return
	}

	// We save the customer but will not display this in the receipt.
	_, err = app.SaveCustomer(txnPtr.FirstName, txnPtr.LastName, txnPtr.Email)
	if err != nil {
		app.errorLog.Println(err)
		app.clientError(w, http.StatusBadRequest)
		return
	}

	txn := models.Transaction{
		Amount:              txnPtr.PaymentAmount,
		Currency:            txnPtr.PaymentCurrency,
		LastFour:            txnPtr.LastFour,
		ExpiryMonth:         txnPtr.ExpiryMonth,
		ExpiryYear:          txnPtr.ExpiryYear,
		BankReturnCode:      txnPtr.BankReturnCode,
		PaymentIntent:       txnPtr.PaymentIntentID,
		PaymentMethod:       txnPtr.PaymentMethodID,
		TransactionStatusID: 2, //cleared
	}

	txnID, err := app.SaveTxn(txn)
	if err != nil {
		app.errorLog.Println(err)
		app.clientError(w, http.StatusBadRequest)
		return
	}
	txnPtr.ID = txnID

	// We are done with the DB, since there is no Order in this case.

	// Dereference the pointer to struct.
	txnData := *txnPtr

	app.Session.Put(r.Context(), "receipt", txnData)
	http.Redirect(w, r, "/receipt", http.StatusSeeOther)

}

func (app *application) PaymentSucceeded(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	txnPtr, err := app.GetTxnData(r)
	if err != nil {
		app.errorLog.Println(err)
		app.clientError(w, http.StatusBadRequest)
		return
	}
	productID, err := strconv.Atoi(r.Form.Get("product_id"))
	if err != nil {
		app.errorLog.Println("widget_id is not an int")
		app.clientError(w, http.StatusBadRequest)
		return
	}

	customerID, err := app.SaveCustomer(txnPtr.FirstName, txnPtr.LastName, txnPtr.Email)
	if err != nil {
		app.errorLog.Println(err)
		app.clientError(w, http.StatusBadRequest)
		return
	}

	txn := models.Transaction{
		Amount:              txnPtr.PaymentAmount,
		Currency:            txnPtr.PaymentCurrency,
		LastFour:            txnPtr.LastFour,
		ExpiryMonth:         txnPtr.ExpiryMonth,
		ExpiryYear:          txnPtr.ExpiryYear,
		BankReturnCode:      txnPtr.BankReturnCode,
		PaymentIntent:       txnPtr.PaymentIntentID,
		PaymentMethod:       txnPtr.PaymentMethodID,
		TransactionStatusID: 2, //cleared
	}

	txnID, err := app.SaveTxn(txn)
	if err != nil {
		app.errorLog.Println(err)
		app.clientError(w, http.StatusBadRequest)
		return
	}
	txnPtr.ID = txnID

	order := models.Order{
		WidgetID:      productID,
		TransactionID: txnID,
		StatusID:      1, // need to check this
		CustomerID:    customerID,
		Quantity:      1, // fixed for the app for now
		Amount:        txnPtr.PaymentAmount,
	}
	// For now, we don't need to use the orderID here:
	_, err = app.SaveOrder(order)
	if err != nil {
		app.errorLog.Println(err)
		app.clientError(w, http.StatusBadRequest)
		return
	}

	// Dereference the pointer to struct.
	txnData := *txnPtr

	app.Session.Put(r.Context(), "receipt", txnData)
	http.Redirect(w, r, "/receipt", http.StatusSeeOther)

}

func (app *application) DisplayReceipt(w http.ResponseWriter, r *http.Request) {
	txnData, ok := app.Session.Get(r.Context(), "receipt").(TransactionData)
	app.infoLog.Println(txnData)
	if !ok {
		app.errorLog.Println("Could not find receipt data in session")
		app.clientError(w, http.StatusBadRequest)
		return
	}
	app.Session.Remove(r.Context(), "receipt")
	data := make(map[string]interface{})
	data["receipt"] = txnData
	if err := app.renderTemplate(w, r, "receipt", &templateData{
		Data: data,
	}); err != nil {
		app.errorLog.Println(err)
	}

}

func (app *application) TestGetWidget(w http.ResponseWriter, r *http.Request) {
	widgetID := 1
	widget, err := app.DB.GetWidget(widgetID)
	if err != nil {
		app.errorLog.Println(err)
		return
	}
	app.infoLog.Println("widget", widget)
}

func (app *application) BuyOneItem(w http.ResponseWriter, r *http.Request) {

	id := chi.URLParam(r, "id")
	widgetID, _ := strconv.Atoi(id)
	app.infoLog.Println("WID:", widgetID)

	widget, err := app.DB.GetWidget(widgetID)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	data := make(map[string]interface{})
	data["widget"] = widget
	tdata := templateData{
		Data: data,
	}

	if err := app.renderTemplate(w, r, "buy-once", &tdata, "stripejs", "stripe-form"); err != nil {
		app.errorLog.Println(err)
	}
}

func (app *application) BronzePlan(w http.ResponseWriter, r *http.Request) {
	widget, err := app.DB.GetWidget(2) // bronze plan
	if err != nil {
		app.errorLog.Println(err)
		return
	}
	data := make(map[string]interface{})
	data["widget"] = widget
	tdata := templateData{
		Data: data,
	}

	if err := app.renderTemplate(w, r, "bronze", &tdata, "stripe-form", "stripejs"); err != nil {
		app.errorLog.Println(err)
	}
}

func (app *application) ReceiptBronze(w http.ResponseWriter, r *http.Request) {
	if err := app.renderTemplate(w, r, "receipt-bronze", nil); err != nil {
		app.errorLog.Println(err)
	}
}

// Authentication
func (app *application) LoginPage(w http.ResponseWriter, r *http.Request) {
	// Make sure we are not logged in while displaying this
	_ = session.Destroy(r.Context())
	_ = session.RenewToken(r.Context())
	// Since we've converted this form to Vue code:
	td := &templateData{}
	if app.vueglue != nil {
		td.VueGlue = app.vueglue
	}
	if err := app.renderTemplate(w, r, "login", td); err != nil {
		app.errorLog.Println(err)
	}
}

func (app *application) ProcessLogin(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	email := r.Form.Get("email")
	password := r.Form.Get("password")

	uid, err := app.DB.Authenticate(email, password)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	session.Put(r.Context(), "userID", uid)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *application) Logout(w http.ResponseWriter, r *http.Request) {
	session.Destroy(r.Context())
	session.RenewToken(r.Context())
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *application) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	if err := app.renderTemplate(w, r, "forgot-password", nil); err != nil {
		app.errorLog.Println(err)
	}
}

func (app *application) ResetPassword(w http.ResponseWriter, r *http.Request) {
	theURL := r.RequestURI
	testURL := fmt.Sprintf("%s%s", app.config.frontend, theURL)
	app.infoLog.Println(testURL)

	signer := urlsigner.Signer{
		Secret: []byte(app.config.secretkey),
	}

	valid := signer.VerifyToken(testURL)

	if !valid {
		app.errorLog.Println("Invalid url - tampering detected")
		app.setFlashAndGoHome(w, r, "Sorry! There was a problem processing your link.", http.StatusSeeOther)
		return
	}

	// Token expires at 60 minutes
	if signer.Expired(testURL, 60) {
		app.setFlashAndGoHome(w, r, "Sorry! Your reset link has expired", http.StatusSeeOther)
		return
	}

	// Make sure the email is hashed as well:
	email := r.URL.Query().Get("email")
	hash, err := signer.GetHashWithSalt(email)
	if err != nil {
		app.errorLog.Println("hasher failed:", err)
		app.clientError(w, http.StatusBadRequest)
		return
	}

	data := make(map[string]interface{})
	data["email"] = email
	data["email_hash"] = hash
	td := templateData{
		Data: data,
	}
	if err := app.renderTemplate(w, r, "reset-password", &td); err != nil {
		app.errorLog.Println(err)
	}
}

func (app *application) PasswordLinkSent(w http.ResponseWriter, r *http.Request) {
	if err := app.renderTemplate(w, r, "link-sent", nil); err != nil {
		app.errorLog.Println(err)
	}
}

// Admin functions

func (app *application) AllSales(w http.ResponseWriter, r *http.Request) {
	if err := app.renderTemplate(w, r, "all-sales", nil); err != nil {
		app.errorLog.Println(err)
	}
}

func (app *application) AllSubscriptions(w http.ResponseWriter, r *http.Request) {
	if err := app.renderTemplate(w, r, "all-subscriptions", nil); err != nil {
		app.errorLog.Println(err)
	}
}

func (app *application) GetSale(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	id, _ := strconv.Atoi(idParam)
	order, err := app.DB.GetSale(id)
	if err != nil {
		app.errorLog.Println(err)
		http.Redirect(w, r, "/", http.StatusNotFound)
		return
	}
	data := make(map[string]interface{})
	data["order"] = order
	td := templateData{
		Data: data,
	}
	if err = app.renderTemplate(w, r, "sale", &td); err != nil {
		app.errorLog.Println(err)
	}
}

func (app *application) GetSubscription(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	id, _ := strconv.Atoi(idParam)
	order, err := app.DB.GetSubscription(id)
	if err != nil {
		app.errorLog.Println(err)
		http.Redirect(w, r, "/", http.StatusNotFound)
		return
	}
	data := make(map[string]interface{})
	data["order"] = order
	td := templateData{
		Data: data,
	}
	if err := app.renderTemplate(w, r, "subscription", &td); err != nil {
		app.errorLog.Println(err)
	}
}

func (app *application) AllUsers(w http.ResponseWriter, r *http.Request) {
	if err := app.renderTemplate(w, r, "all-users", nil); err != nil {
		app.errorLog.Println(err)
	}
}

func (app *application) ShowUser(w http.ResponseWriter, r *http.Request) {
	uid, err := strconv.Atoi(chi.URLParam(r, "id"))
	user, err := app.DB.GetUserByID(uid)
	data := make(map[string]interface{})
	data["user"] = user
	td := templateData{
		Data: data,
	}
	if err = app.renderTemplate(w, r, "show-user", &td); err != nil {
		app.errorLog.Println(err)
	}
}

func (app *application) EditUser(w http.ResponseWriter, r *http.Request) {
	uid, _ := strconv.Atoi(chi.URLParam(r, "id"))
	user, err := app.DB.GetUserByID(uid)
	data := make(map[string]interface{})
	data["user"] = user
	td := templateData{
		Data: data,
	}
	if err = app.renderTemplate(w, r, "new-user", &td); err != nil {
		app.errorLog.Println(err)
	}
}

func (app *application) NewUserForm(w http.ResponseWriter, r *http.Request) {
	if err := app.renderTemplate(w, r, "new-user", nil); err != nil {
		app.errorLog.Println(err)
	}
}
