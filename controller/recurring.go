package controller

import (
	"encoding/json"
	"net/http"

	adyen "github.com/zhutik/adyen-api-go"
)

// PerformRecurringList - get list of saved payment methods for a given shopper
func PerformRecurringList(instance *adyen.Adyen, w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	req := &adyen.RecurringDetailsRequest{
		MerchantAccount:  instance.MerchantAccount,
		ShopperReference: r.Form.Get("shopperReference"),
	}

	g, err := instance.Recurring().ListRecurringDetails(req)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response, err := json.Marshal(g)

	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
}
