/*
	export ADYEN_CLIENT_TOKEN="YOUR_ADYEN_ENCRYPTED_URL"
	export ADYEN_USERNAME="YOUR_ADYEN_API_USERNAME"
	export ADYEN_PASSWORD="YOUR_API_PASSWORD"
	export ADYEN_ACCOUNT="YOUR_MERCHANT_ACCOUNT"

	# API settings for Adyen Hosted Payment pages
	export ADYEN_HMAC="YOUR_HMAC_KEY"
	export ADYEN_SKINCODE="YOUR_SKIN_CODE"
	export ADYEN_SHOPPER_LOCALE="en_GB"
*/

package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"./controller"
	"./logger"

	"github.com/ernesto-jimenez/httplogger"
	"github.com/gorilla/mux"
	"github.com/zhutik/adyen-api-go"
)

// TemplateConfig for HTML template
type TemplateConfig struct {
	EncURL string
	Time   string
}

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
		adyen.Testing,
		os.Getenv("ADYEN_USERNAME"),
		os.Getenv("ADYEN_PASSWORD"),
		adyen.WithTransport(httplogger.NewLoggedTransport(http.DefaultTransport, logger.NewLogger())),
	)

	instance.Currency = "EUR"
	instance.MerchantAccount = os.Getenv("ADYEN_ACCOUNT")

	return instance
}

func initAdyenHPP() *adyen.Adyen {
	instance := adyen.NewWithHMAC(
		adyen.Testing,
		os.Getenv("ADYEN_USERNAME"),
		os.Getenv("ADYEN_PASSWORD"),
		os.Getenv("ADYEN_HMAC"),
		adyen.WithTransport(httplogger.NewLoggedTransport(http.DefaultTransport, logger.NewLogger())),
	)

	instance.Currency = "EUR"
	instance.MerchantAccount = os.Getenv("ADYEN_ACCOUNT")

	return instance
}

/**
 * Show Adyen Payment form
 */
func showForm(w http.ResponseWriter, r *http.Request) {
	instance := initAdyen()

	now := time.Now()
	cwd, _ := os.Getwd()

	config := TemplateConfig{
		EncURL: instance.ClientURL(os.Getenv("ADYEN_CLIENT_TOKEN")),
		Time:   now.Format(time.RFC3339),
	}

	t := template.Must(template.ParseGlob(filepath.Join(cwd, "./templates/*")))
	err := t.ExecuteTemplate(w, "indexPage", config)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func performDirectoryLookup(w http.ResponseWriter, r *http.Request) {
	instance := initAdyenHPP()

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

func performHpp(w http.ResponseWriter, r *http.Request) {
	instance := initAdyenHPP()

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

func performRecurringList(w http.ResponseWriter, r *http.Request) {
	instance := initAdyen()

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

func performPayment(w http.ResponseWriter, r *http.Request) {
	controller.PerformPayment(initAdyen(), w, r)
}

func performCapture(w http.ResponseWriter, r *http.Request) {
	controller.PerformCapture(initAdyen(), w, r)
}

func performCancel(w http.ResponseWriter, r *http.Request) {
	controller.PerformCancel(initAdyen(), w, r)
}

func performRefund(w http.ResponseWriter, r *http.Request) {
	controller.PerformRefund(initAdyen(), w, r)
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

	cwd, err := os.Getwd()
	if err != nil {
		panic("Can't read current working directory")
	}

	r := mux.NewRouter()

	r.HandleFunc("/", showForm)
	r.HandleFunc("/perform_payment", performPayment)
	r.HandleFunc("/perform_capture", performCapture)
	r.HandleFunc("/perform_cancel", performCancel)
	r.HandleFunc("/perform_lookup", performDirectoryLookup)
	r.HandleFunc("/perform_hpp", performHpp)
	r.HandleFunc("/perform_refund", performRefund)
	r.HandleFunc("/perform_recurring_list", performRecurringList)

	s := http.StripPrefix("/static/", http.FileServer(http.Dir(cwd+"/static/")))
	r.PathPrefix("/static/").Handler(s)

	http.Handle("/", r)

	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
