package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/stripe/stripe-go/v72"
	portalsession "github.com/stripe/stripe-go/v72/billingportal/session"
	"github.com/stripe/stripe-go/v72/checkout/session"
	"github.com/stripe/stripe-go/v72/webhook"
)

type errResp struct {
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
}

func writeJSON(w http.ResponseWriter, v interface{}, err error) {
	var respVal interface{}
	if err != nil {
		msg := err.Error()
		var serr *stripe.Error
		if errors.As(err, &serr) {
			msg = serr.Msg
		}
		w.WriteHeader(http.StatusBadRequest)
		var e errResp
		e.Error.Message = msg
		respVal = e
	} else {
		respVal = v
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(respVal); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("json.NewEncoder.Encode: %v", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if _, err := io.Copy(w, &buf); err != nil {
		log.Printf("io.Copy: %v", err)
		return
	}
}

func (app *App) upgrade(user *User, w http.ResponseWriter, r *http.Request) {
	if user.Advanced {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}

	renderTemplate(w, "upgrade", &PageView{PageTitle: "UPGRADE", User: *user})
}

func (app *App) handleCanceled(user *User, w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "canceled", &PageView{PageTitle: "STRIPE_CANCELED", User: *user})
}

func (app *App) handleManageAccount(user *User, w http.ResponseWriter, r *http.Request) {
	if user.Advanced == false {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}

	renderTemplate(w, "manage", &PageView{
		PageTitle: "STRIPE_MANAGE",
		User:      *user,
	})
}

func (app *App) handleConfig(user *User, w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	fmt.Println("config")

	writeJSON(w, struct {
		PublishableKey string `json:"publishableKey"`
		BasicPrice     string `json:"basicPrice"`
		ProPrice       string `json:"proPrice"`
	}{
		PublishableKey: app.config.StripeKey,
		BasicPrice:     app.config.StripeAdvancePriceId,
		ProPrice:       app.config.StripeAdvancePriceId,
	}, nil)
}

func (app *App) handleCreateCheckoutSession(user *User, w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	fmt.Println("create checkout")

	r.ParseForm()
	priceId := r.PostFormValue("priceId")
	params := &stripe.CheckoutSessionParams{
		SuccessURL:    stripe.String(app.config.StripeReturnURL + "/checkout-session?session_id={CHECKOUT_SESSION_ID}"),
		CancelURL:     stripe.String(app.config.StripeReturnURL + "/canceled"),
		Mode:          stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		CustomerEmail: &(user.Email),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(priceId),
				Quantity: stripe.Int64(1),
			},
		},
		// AutomaticTax: &stripe.CheckoutSessionAutomaticTaxParams{Enabled: stripe.Bool(true)},
	}

	s, err := session.New(params)
	if err != nil {
		writeJSON(w, nil, err)
		return
	}

	http.Redirect(w, r, s.URL, http.StatusSeeOther)
}

func (app *App) handleCheckoutSession(user *User, w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	sessionID := r.URL.Query().Get("session_id")
	s, err := session.Get(sessionID, nil)
	if err != nil {
		log.Printf("CheckoutSession: %v\n", err)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	fmt.Println(s.Customer.ID)

	app.datastore.upgradeUser(user.Id, sessionID, s.Customer.ID)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *App) handleWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	fmt.Println("webhook")

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Printf("ioutil.ReadAll: %v", err)
		return
	}

	event, err := webhook.ConstructEvent(b, r.Header.Get("Stripe-Signature"), app.config.StripeSecret)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Printf("webhook.ConstructEvent: %v", err)
		return
	}

	if event.Type != "checkout.session.completed" {
		return
	}
}

func (app *App) handleCustomerPortal(user *User, w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	r.ParseForm()
	sessionID := r.PostFormValue("sessionId")[0:]

	// For demonstration purposes, we're using the Checkout session to retrieve the customer ID.
	// Typically this is stored alongside the authenticated user in your database.
	s, err := session.Get(sessionID, nil)
	if err != nil {
		writeJSON(w, nil, err)
		return
	}

	// The URL to which the user is redirected when they are done managing
	// billing in the portal.
	returnURL := app.config.StripeReturnURL

	params := &stripe.BillingPortalSessionParams{
		Customer:  stripe.String(s.Customer.ID),
		ReturnURL: stripe.String(returnURL),
	}
	ps, _ := portalsession.New(params)

	http.Redirect(w, r, ps.URL, http.StatusSeeOther)
}
