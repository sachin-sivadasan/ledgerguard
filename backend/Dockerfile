# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /build

# Install dependencies
RUN apk add --no-cache git ca-certificates

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /build/server ./cmd/server

# Runtime stage
FROM alpine:3.19

WORKDIR /app

# Install ca-certificates for HTTPS calls
RUN apk add --no-cache ca-certificates tzdata

# Copy binary from builder
COPY --from=builder /build/server /app/server

# Copy migrations
COPY --from=builder /build/migrations /app/migrations

# Set environment
ENV FIREBASE_CREDENTIALS_FILE=/app/firebase-credentials.json
ENV DB_MIGRATIONS_PATH=/app/migrations

EXPOSE 8080

ENTRYPOINT ["/app/server"]
