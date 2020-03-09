FROM golang:1.13

WORKDIR /go/src/app

COPY . .

RUN go get -u github.com/zhutik/adyen-api-go
RUN go build -v ./...

EXPOSE 8080

CMD ["go", "run", "main.go"]
