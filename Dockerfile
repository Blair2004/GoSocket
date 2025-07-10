# Multi-stage build for smaller production image
FROM golang:1.21-alpine AS builder

# Install git for go modules
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binaries
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o socket-server main.go
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o socket cmd/cli/main.go

# Production stage
FROM alpine:latest

# Install ca-certificates for SSL/TLS
RUN apk --no-cache add ca-certificates

# Create non-root user
RUN adduser -D -s /bin/sh socketuser

# Set working directory
WORKDIR /app

# Copy binaries from builder stage
COPY --from=builder /app/socket-server .
COPY --from=builder /app/socket .

# Copy web assets
COPY --from=builder /app/web ./web

# Change ownership to non-root user
RUN chown -R socketuser:socketuser /app

# Switch to non-root user
USER socketuser

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ./socket --server http://localhost:8080 health || exit 1

# Run the server
CMD ["./socket-server"]
