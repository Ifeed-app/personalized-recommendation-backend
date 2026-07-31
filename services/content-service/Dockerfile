FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY proto ./proto
COPY services/content-service ./services/content-service

WORKDIR /app/services/content-service

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -o app ./cmd

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/services/content-service/app /app/app

RUN chmod +x /app/app

CMD ["/app/app"]