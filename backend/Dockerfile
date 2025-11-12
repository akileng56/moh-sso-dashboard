# ======================
# Stage 1: Build
# ======================
FROM golang:1.25-alpine AS builder

# Set up environment
ENV CGO_ENABLED=0 \
  GOOS=linux \
  GOARCH=amd64

WORKDIR /app

# Copy go.mod and go.sum first for dependency caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source
COPY . .

# Build binary with optimizations
RUN go build -ldflags="-s -w" -o sso-server ./cmd/server

# ======================
# Stage 2: Runtime
# ======================
FROM alpine:3.19

# Add CA certificates for HTTPS support
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/sso-server .

# Use a non-root user for security
RUN adduser -D appuser
USER appuser

EXPOSE 9000

# Start the server
CMD ["./sso-server"]
