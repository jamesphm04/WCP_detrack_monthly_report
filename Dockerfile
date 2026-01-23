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

# Build binary
RUN go build -o wcp_detrack_monthly_report ./cmd/main.go

# --- Stage 2: Create a small runtime image ---
FROM alpine:latest

# Install certificates and timezone data
RUN apk add --no-cache ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/wcp_detrack_monthly_report .

# Copy .env file 
COPY .env .

# Command to run the binary
ENTRYPOINT ["./wcp_detrack_monthly_report"]