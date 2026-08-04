FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY proto ./proto
COPY shared ./shared
COPY services/upload-service ./services/upload-service

WORKDIR /app/services/upload-service

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -o app ./cmd

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/services/upload-service/app /app/app
RUN chmod +x /app/app

EXPOSE 8050

CMD ["/app/app"]