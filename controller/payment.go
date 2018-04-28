package controller

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	adyen "github.com/zhutik/adyen-api-go"
)

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
