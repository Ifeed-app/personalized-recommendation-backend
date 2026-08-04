FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY proto ./proto
COPY shared ./shared
COPY services/user-service ./services/user-service

WORKDIR /app/services/user-service
ENV GOWORK=off

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o app ./cmd

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/services/user-service/app /app/app
RUN chmod +x /app/app

EXPOSE 8060

CMD ["/app/app"]