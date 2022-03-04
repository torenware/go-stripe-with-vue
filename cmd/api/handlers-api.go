package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/torenware/go-stripe/internal/cards"
)

type stripePayload struct {
	Currency string `json:"currency"`
	Amount   string `json:"amount"`
}

type jsonResponse struct {
	OK      bool   `json:"ok"`
	Message string `json:"message,omitempty"`
	Content string `json:"content,omitempty"`
	ID      int    `json:"id,omitempty"`
}

// let payload = {
// 	plan: '{{$widget.PlanID}}',
// 	payment_method: result.paymentMethod.id,
// 	email: document.getElementById("email").value,
// 	last_four: result.paymentMethod.card.last4,
//   };

type subscriptionPayload struct {
	PlanID        string `json:"plan"`
	PaymentMethod string `json:"payment_method"`
	Email         string `json:"email"`
	LastFour      string `json:"last_four"`
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
	// Convert the amount to an int
	amount, err := strconv.Atoi(payload.Amount)
	if err != nil {
		// again, replace later
		app.errorLog.Println(err)
		return
	}

	card := cards.Card{
		Secret:   app.config.stripe.secret,
		Key:      app.config.stripe.key,
		Currency: payload.Currency,
	}

	okay := true // optimism

	pi, msg, err := card.Charge(payload.Currency, amount)
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
	var payload subscriptionPayload

	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	test, err := json.MarshalIndent(payload, "", "    ")
	app.infoLog.Println(string(test))

	// stub
	j := jsonResponse{
		OK:      true,
		Message: "Pong",
	}
	out, err := json.MarshalIndent(j, "", "  ")
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(out)

}
