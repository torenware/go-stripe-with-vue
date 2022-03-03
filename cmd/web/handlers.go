package main

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/torenware/go-stripe/internal/cards"
	"github.com/torenware/go-stripe/internal/models"
)

// The clientError helper sends a specific status code and corresponding description
// to the user. We'll use this later in the book to send responses like 400 "Bad
// Request" when there's a problem with the request that the user sent.
func (app *application) clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

func (app *application) HomePage(w http.ResponseWriter, r *http.Request) {
	if err := app.renderTemplate(w, r, "home", nil); err != nil {
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
	err := r.ParseForm()
	if err != nil {
		return nil, err
	}

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

func (app *application) PaymentSucceeded(w http.ResponseWriter, r *http.Request) {
	app.infoLog.Println("Form submission")
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	txnPtr, err := app.GetTxnData(r)
	if err != nil {
		app.errorLog.Println("payment amount is not an int")
		app.clientError(w, http.StatusBadRequest)
		return
	}
	// Form was parsed in GetTxnData
	productID, err := strconv.Atoi(r.Form.Get("product_id"))
	if err != nil {
		app.errorLog.Println("payment amount is not an int")
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

	// data := make(map[string]interface{})
	// data["cardholder"] = cardHolder
	// data["email"] = email
	// data["pi"] = paymentIntent
	// data["pm"] = paymentMethod
	// data["pa"] = paymentAmount
	// data["pc"] = paymentCurrency
	// data["last_four"] = lastFour
	// data["expiry_month"] = expiryMonth
	// data["expiry_year"] = expiryYear
	// data["bank_return_code"] = bankReturnCode
	// data["order_id"] = orderID

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
