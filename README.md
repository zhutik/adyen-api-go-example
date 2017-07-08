# [WIP] Adyen API for Go Example Playground

## Install

```
go get github.com/zhutik/adyen-api-go
```

![Alt text](./screenshots/authorize.png "Playground example")

## Supported API Calls
* Authorize
* Authorize encrypted (default)
* Capture
* Cancel

## Next
* Refund
* Refund or Cancel
* Notifications


## Configuration

### Expose your settings for Adyen API configuration.

```main.go``` script will use those variables to communicate with API

```
$ export ADYEN_CLIENT_TOKEN="YOUR_ADYEN_CLIENT_TOKEN"
$ export ADYEN_USERNAME="YOUR_ADYEN_API_USERNAME"
$ export ADYEN_PASSWORD="YOUR_API_PASSWORD"
$ export ADYEN_ACCOUNT="YOUR_MERCHANT_ACCOUNT"
```

Or, modify ```.default.env.template```

```
cp .default.env.template .default.env

# modify/change .default.env and put your credentials

source .default.env
```

Settings explanation:
* ADYEN_CLIENT_TOKEN - Library token in Adyen, used to load external JS file from Adyen to validate Credit Card information
* ADYEN_USERNAME - Adyen API username, usually starts with ws@
* ADYEN_PASSWORD - Adyen API password for username
* ADYEN_ACCOUNT - Selected Merchant Account

## Hosted Payment Pages

![Alt text](./screenshots/hosted_payment_methods.png "Playground example")

update your configuration and make sure you specify additional parameters

```
# API settings for Adyen Hosted Payment pages
$ export ADYEN_HMAC="YOUR_HMAC_KEY"
$ export ADYEN_SKINCODE="YOUR_SKINCODE_ID"
$ export ADYEN_SHOPPER_LOCALE="YOUR_SHOPPER_LOCALE"
```

## Run with Docker-compose

Note: Expose your configuration (as shown above)

```
$ docker-compose up

# or 

$ docker-compose up -d
```

Open ```http://localhost:8080``` in your browser


## Run example without Docker

```
# Install dependencies
$ go get -d -v ./...
$ go install -v ./...

$ go run main.go
```

### Perform payments

Open ```http://localhost:8080``` in your browser
Put credit card information.

Test credit cards could be found https://docs.adyen.com/support/integration#testcardnumbers

## Contribute

Please check initial library repository https://github.com/zhutik/adyen-api-go
