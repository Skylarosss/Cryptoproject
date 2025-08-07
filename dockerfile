FROM golang:alpine AS builder

WORKDIR /build

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN go build -o cryptoproject ./cmd/app/main.go

FROM alpine:latest

RUN mkdir -p /app/config

COPY --from=builder /build/cryptoproject /app/

COPY config/cfg.yaml /app/config/cfg.yaml

WORKDIR /app

CMD ["/app/cryptoproject"]