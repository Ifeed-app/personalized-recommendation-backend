FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY services/email-worker ./services/email-worker

WORKDIR /app/services/email-worker

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux \
    go build -o app ./cmd

FROM alpine:latest

WORKDIR /app

COPY --from=builder \
    /app/services/email-worker/app \
    /app/app

RUN chmod +x /app/app

CMD ["/app/app"]