FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY proto ./proto
COPY services/interaction-service ./services/interaction-service

WORKDIR /app/services/interaction-service

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -o app ./cmd

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/services/interaction-service/app /app/app

RUN chmod +x /app/app

CMD ["/app/app"]