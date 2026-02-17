# --- Stage 1: Build Go binary ---
FROM golang:1.25.5-alpine AS builder

# Install git, ssh, and ca-certificates. Sometimes installing dependency needs git
RUN apk add --no-cache git openssh ca-certificates bash

WORKDIR /app

# Copy go.mod and go.sum first (dependency caching)
COPY go.mod go.sum ./
RUN go mod download

# Copy rest of the code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/main ./cmd/main.go

# --- Stage 2: Create a small runtime image ---
FROM alpine:latest

# Install certificates and timezone data
RUN apk add --no-cache ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/bin/main .

# Create logs directory and set permissions BEFORE switching to non-root user
RUN mkdir -p /app/logs && \
    adduser -D -u 1000 appuser && \
    chown -R appuser:appuser /app

CMD ["./main"]