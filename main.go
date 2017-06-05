/*
	export ADYEN_CLIENT_TOKEN="YOUR_ADYEN_ENCRYPTED_URL"
	export ADYEN_USERNAME="YOUR_ADYEN_API_USERNAME"
	export ADYEN_PASSWORD="YOUR_API_PASSWORD"
	export ADYEN_ACCOUNT="YOUR_MERCHANT_ACCOUNT"
*/

package main

import (
	"encoding/json"
	"fmt"
	adyen "github.com/zhutik/adyen-api-go"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// TempateConfig for HTML template
type TempateConfig struct {
	EncURL string
	Time   string
}

var Logger *log.Logger

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

// initAdyen init Adyen API instance
func initAdyen() *adyen.Adyen {
	instance := adyen.New(
		os.Getenv("ADYEN_USERNAME"),
		os.Getenv("ADYEN_PASSWORD"),
		os.Getenv("ADYEN_CLIENT_TOKEN"),
		os.Getenv("ADYEN_ACCOUNT"),
	)

	Logger = log.New(os.Stdout, "Adyen Playground: ", log.Ldate|log.Ltime|log.Lshortfile)

	instance.SetCurrency("EUR")
	instance.AttachLogger(Logger)

	return instance
}

/**
 * Show Adyen Payment form
 */
func showForm(w http.ResponseWriter, r *http.Request) {
	instance := adyen.New(
		os.Getenv("ADYEN_USERNAME"),
		os.Getenv("ADYEN_PASSWORD"),
		os.Getenv("ADYEN_CLIENT_TOKEN"),
		os.Getenv("ADYEN_ACCOUNT"),
	)
	now := time.Now()
	cwd, _ := os.Getwd()

	config := TempateConfig{
		EncURL: instance.ClientURL(),
		Time:   now.Format(time.RFC3339),
	}

	t := template.Must(template.ParseGlob(filepath.Join(cwd, "./templates/*")))
	err := t.ExecuteTemplate(w, "indexPage", config)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

/**
 * Handle post request and perform payment authosization
 */
func performPayment(w http.ResponseWriter, r *http.Request) {
	instance := initAdyen()

	r.ParseForm()

	rand.Seed(time.Now().UTC().UnixNano())

	var g *adyen.AuthoriseResponse
	var err error

	// Form was submited with encrypted data
	if len(r.Form.Get("adyen-encrypted-data")) > 0 {
		req := &adyen.AuthoriseEncrypted{
			Amount:          &adyen.Amount{Value: 1000, Currency: instance.Currency},
			MerchantAccount: os.Getenv("ADYEN_ACCOUNT"),
			AdditionalData:  &adyen.AdditionalData{Content: r.Form.Get("adyen-encrypted-data")},
			Reference:       "DE-100" + randomString(6), // order number or some bussiness reference
		}

		g, err = instance.Payment().AuthoriseEncrypted(req)
	} else {
		month, _ := strconv.Atoi(r.Form.Get("expiryMonth"))
		year, _ := strconv.Atoi(r.Form.Get("expiryYear"))
		cvc, _ := strconv.Atoi(r.Form.Get("cvc"))

		req := &adyen.Authorise{
			Card: &adyen.Card{
				Number:      r.Form.Get("number"),
				ExpireMonth: month,
				ExpireYear:  year,
				HolderName:  r.Form.Get("holderName"),
				Cvc:         cvc,
			},
			Amount:          &adyen.Amount{Value: 1000, Currency: instance.Currency},
			MerchantAccount: os.Getenv("ADYEN_ACCOUNT"),
			Reference:       "DE-100" + randomString(6), // order number or some bussiness reference
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

func performCapture(w http.ResponseWriter, r *http.Request) {
	instance := initAdyen()

	r.ParseForm()

	amount, err := strconv.ParseFloat(r.Form.Get("amount"), 32)

	if err != nil {
		http.Error(w, "Failed! Can not convert amount to float", http.StatusInternalServerError)
		return
	}

	req := &adyen.Capture{
		ModificationAmount: &adyen.Amount{Value: float32(amount), Currency: instance.Currency},
		MerchantAccount:    os.Getenv("ADYEN_ACCOUNT"),       // Merchant Account setting
		Reference:          r.Form.Get("reference"),          // order number or some bussiness reference
		OriginalReference:  r.Form.Get("original-reference"), // PSP reference that came as authosization results
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

func performCancel(w http.ResponseWriter, r *http.Request) {
	instance := initAdyen()

	r.ParseForm()

	req := &adyen.Cancel{
		Reference:         r.Form.Get("reference"),          // order number or some bussiness reference
		MerchantAccount:   os.Getenv("ADYEN_ACCOUNT"),       // Merchant Account setting
		OriginalReference: r.Form.Get("original-reference"), // PSP reference that came as authosization result
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

func main() {
	fmt.Println("Checking environment variables...")

	if len(os.Getenv("ADYEN_USERNAME")) == 0 ||
		len(os.Getenv("ADYEN_PASSWORD")) == 0 ||
		len(os.Getenv("ADYEN_CLIENT_TOKEN")) == 0 ||
		len(os.Getenv("ADYEN_ACCOUNT")) == 0 {
		panic("Some of the required varibles are missing or empty.\nPlease make sure\nADYEN_USERNAME\nADYEN_PASSWORD\nADYEN_CLIENT_TOKEN\nADYEN_ACCOUNT\nare set as environment variables")
	}

	port := 8080

	if len(os.Getenv("APPLICATION_PORT")) != 0 {
		port, _ = strconv.Atoi(os.Getenv("APPLICATION_PORT"))
	}

	fmt.Println(fmt.Sprintf("Start listening connections on port %d...", port))

	http.HandleFunc("/", showForm)
	http.HandleFunc("/perform_payment", performPayment)
	http.HandleFunc("/perform_capture", performCapture)
	http.HandleFunc("/perform_cancel", performCancel)
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
