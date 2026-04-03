FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o app ./cmd

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/app /app/app
RUN chmod +x /app/app

EXPOSE 8050

CMD ["/app/app"]