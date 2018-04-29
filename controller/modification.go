package controller

import (
	"encoding/json"
	"net/http"
	"strconv"

	adyen "github.com/zhutik/adyen-api-go"
)

// PerformCapture - performs capture request for given authorization
func PerformCapture(instance *adyen.Adyen, w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	amount, err := strconv.ParseFloat(r.Form.Get("amount"), 32)

	if err != nil {
		http.Error(w, "Failed! Can not convert amount to float", http.StatusInternalServerError)
		return
	}

	req := &adyen.Capture{
		ModificationAmount: &adyen.Amount{Value: float32(amount), Currency: instance.Currency},
		MerchantAccount:    instance.MerchantAccount,         // Merchant Account setting
		Reference:          r.Form.Get("reference"),          // order number or some business reference
		OriginalReference:  r.Form.Get("original-reference"), // PSP reference that came as authorization results
	}

	g, err := instance.Modification().Capture(req)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response, err := json.Marshal(g)

	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
}

// PerformCancel - cancel given authorized transaction
func PerformCancel(instance *adyen.Adyen, w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	req := &adyen.Cancel{
		Reference:         r.Form.Get("reference"),          // order number or some business reference
		MerchantAccount:   instance.MerchantAccount,         // Merchant Account setting
		OriginalReference: r.Form.Get("original-reference"), // PSP reference that came as authorization result
	}

	g, err := instance.Modification().Cancel(req)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response, err := json.Marshal(g)

	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
}

// PerformRefund - refunds given captured transaction
func PerformRefund(instance *adyen.Adyen, w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	amount, err := strconv.ParseFloat(r.Form.Get("amount"), 32)

	if err != nil {
		http.Error(w, "Failed! Can not convert amount to float", http.StatusInternalServerError)
		return
	}

	req := &adyen.Refund{
		ModificationAmount: &adyen.Amount{Value: float32(amount), Currency: instance.Currency},
		Reference:          r.Form.Get("reference"),          // order number or some business reference
		MerchantAccount:    instance.MerchantAccount,         // Merchant Account setting
		OriginalReference:  r.Form.Get("original-reference"), // PSP reference that came as authorization result
	}

	g, err := instance.Modification().Refund(req)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response, err := json.Marshal(g)

	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
}
