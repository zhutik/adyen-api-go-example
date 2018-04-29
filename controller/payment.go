package controller

import (
	"encoding/json"
	"html/template"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	adyen "github.com/zhutik/adyen-api-go"
)

func randInt(min int, max int) int {
	return min + rand.Intn(max-min)
}

func randomString(l int) string {
	bytes := make([]byte, l)
	for i := 0; i < l; i++ {
		bytes[i] = byte(randInt(65, 90))
	}
	return string(bytes)
}

// PerformPayment - Handle post request and perform payment authorization
func PerformPayment(instance *adyen.Adyen, w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	rand.Seed(time.Now().UTC().UnixNano())

	var g *adyen.AuthoriseResponse
	var err error

	amount, err := strconv.ParseFloat(r.Form.Get("amount"), 32)

	if err != nil {
		http.Error(w, "Failed! Can not convert amount to float", http.StatusInternalServerError)
		return
	}

	reference := r.Form.Get("reference")

	// multiple value by 100, as specified in Adyen documentation
	adyenAmount := float32(amount) * 100

	// Form was submitted with encrypted data
	if len(r.Form.Get("adyen-encrypted-data")) > 0 {
		req := &adyen.AuthoriseEncrypted{
			Amount:           &adyen.Amount{Value: adyenAmount, Currency: instance.Currency},
			MerchantAccount:  instance.MerchantAccount,
			AdditionalData:   &adyen.AdditionalData{Content: r.Form.Get("adyen-encrypted-data")},
			ShopperReference: r.Form.Get("shopperReference"),
			Reference:        reference, // order number or some business reference
		}

		if len(r.Form.Get("is_recurring")) > 0 {
			req.Recurring = &adyen.Recurring{Contract: adyen.RecurringPaymentRecurring}
		}

		g, err = instance.Payment().AuthoriseEncrypted(req)
	} else {
		req := &adyen.Authorise{
			Card: &adyen.Card{
				Number:      r.Form.Get("number"),
				ExpireMonth: r.Form.Get("expiryMonth"),
				ExpireYear:  r.Form.Get("expiryYear"),
				HolderName:  r.Form.Get("holderName"),
				Cvc:         r.Form.Get("cvc"),
			},
			Amount:           &adyen.Amount{Value: adyenAmount, Currency: instance.Currency},
			MerchantAccount:  instance.MerchantAccount,
			Reference:        reference, // order number or some business reference
			ShopperReference: r.Form.Get("shopperReference"),
		}

		g, err = instance.Payment().Authorise(req)
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response, err := json.Marshal(g)

	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
}

// PerformDirectoryLookup - directory look up request
func PerformDirectoryLookup(instance *adyen.Adyen, w http.ResponseWriter, r *http.Request) {
	timeIn := time.Now().Local().Add(time.Minute * time.Duration(60))

	req := &adyen.DirectoryLookupRequest{
		CurrencyCode:      instance.Currency,
		MerchantAccount:   instance.MerchantAccount,
		PaymentAmount:     1000,
		SkinCode:          os.Getenv("ADYEN_SKINCODE"),
		MerchantReference: "DE-100" + randomString(6),
		SessionsValidity:  timeIn.Format(time.RFC3339),
		CountryCode:       "NL",
	}

	g, err := instance.Payment().DirectoryLookup(req)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	cwd, _ := os.Getwd()
	t := template.Must(template.ParseGlob(filepath.Join(cwd, "./templates/*")))
	err = t.ExecuteTemplate(w, "hpp_payment_methods", g)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// PerformHpp perform HPP request and redirect customer to 3rd patry URL
func PerformHpp(instance *adyen.Adyen, w http.ResponseWriter, r *http.Request) {
	timeIn := time.Now().Local().Add(time.Minute * time.Duration(60))

	// 5 days
	shipTime := time.Now().Local().Add(time.Hour * 24 * time.Duration(5))

	req := &adyen.SkipHppRequest{
		MerchantReference: "DE-100" + randomString(6),
		PaymentAmount:     1000,
		CurrencyCode:      instance.Currency,
		ShipBeforeDate:    shipTime.Format(time.RFC3339),
		SkinCode:          os.Getenv("ADYEN_SKINCODE"),
		MerchantAccount:   instance.MerchantAccount,
		ShopperLocale:     "en_GB",
		SessionsValidity:  timeIn.Format(time.RFC3339),
		CountryCode:       "NL",
		BrandCode:         "ideal",
		IssuerID:          "1121",
	}

	url, err := instance.Payment().GetHPPRedirectURL(req)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}
