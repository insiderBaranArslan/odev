# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app
RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /banking-api ./cmd/server

# Runtime stage
FROM alpine:3.20

RUN apk add --no-cache ca-certificates
WORKDIR /app

COPY --from=builder /banking-api /app/banking-api
COPY migrations /app/migrations

EXPOSE 8080
CMD ["/app/banking-api"]
