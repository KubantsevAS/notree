FROM golang:1.26.1-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /backend ./cmd/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /backend /app/backend

COPY --from=builder /app/config /app/config

EXPOSE 8080

CMD ["/app/backend"]