FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY proto ./proto
COPY shared ./shared
COPY services/recommendation-service ./services/recommendation-service

WORKDIR /app/services/recommendation-service

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -o app ./cmd

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/services/recommendation-service/app /app/app

RUN chmod +x /app/app

CMD ["/app/app"]