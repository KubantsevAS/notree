FROM golang:1.26.2-alpine3.23 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /backend ./cmd/main.go

FROM alpine:3.23.3
WORKDIR /app
COPY --from=builder /backend /app/backend
COPY --from=builder /app/config /app/config
EXPOSE 8080
CMD ["/app/backend"]