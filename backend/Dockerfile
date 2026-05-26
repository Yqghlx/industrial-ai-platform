# Backend Dockerfile - Multi-stage build
# Industrial AI Agent Platform

# Build stage
FROM golang:1.25-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Allow Go to auto-download required toolchain
ENV GOTOOLCHAIN=auto

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Debug: Print migration content to verify
RUN echo "=== Migration 000002 content ===" && cat migrations/000002_timescaledb.up.sql

# Build binary with optimizations (force rebuild embed)
ARG TARGETOS
ARG TARGETARCH
RUN go clean -cache && GOTOOLCHAIN=auto CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -a -ldflags="-w -s -extldflags '-static'" \
    -o /build/bin/server .

# Runtime stage
FROM alpine:3.19

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata curl

# Create non-root user
RUN addgroup -g 1000 -S appgroup && \
    adduser -u 1000 -S appuser -G appgroup

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/bin/server .

# Copy migrations (if needed)
COPY --from=builder /build/migrations ./migrations

# Set ownership
RUN chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# Default environment
ENV PORT=8080
ENV GIN_MODE=release

# Run binary
ENTRYPOINT ["./server"]