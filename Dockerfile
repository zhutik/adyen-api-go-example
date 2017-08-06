FROM golang:1.8

WORKDIR /go/src/app

COPY . .

RUN go-wrapper download && \
    go-wrapper install

EXPOSE 8080

CMD ["go-wrapper", "run"] # ["main.go"]
