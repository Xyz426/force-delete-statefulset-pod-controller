FROM golang:1.24-alpine AS builder

WORKDIR /app
RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /custom_controller main.go

FROM alpine:3.18

WORKDIR /
COPY --from=builder /custom_controller /custom_controller

ENTRYPOINT ["/custom_controller"]
