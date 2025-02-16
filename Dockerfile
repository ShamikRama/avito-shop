FROM golang:1.24-alpine

WORKDIR /avito-shop

COPY go.mod go.sum ./

RUN go mod download

COPY cmd/ ./cmd/

COPY internal/ ./internal/

COPY migrations/ ./migrations/

COPY .env .

RUN go build -o avitoservice ./cmd/

RUN go install github.com/pressly/goose/v3/cmd/goose@latest

RUN apk add --no-cache curl && \
    curl -L https://raw.githubusercontent.com/vishnubob/wait-for-it/master/wait-for-it.sh -o /usr/local/bin/wait-for-it && \
    chmod +x /usr/local/bin/wait-for-it

RUN apk add --no-cache bash

ENV PATH="/usr/local/bin:${PATH}"

ENV CONFIG_PATH=/avito-shop/.env

CMD ["./avitoservice"]