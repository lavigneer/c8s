# Multi-stage Dockerfile for C8S components
# Builds controller and webhook binaries

# Build stage
FROM golang:1.26-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make ca-certificates

# Set working directory
WORKDIR /workspace

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY cmd/ cmd/
COPY pkg/ pkg/
COPY hack/ hack/
COPY PROJECT ./
COPY Makefile ./

# Build deployed components
# Use -trimpath to remove build paths (avoid -a flag to leverage layer caching)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -trimpath -o bin/controller ./cmd/controller
RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -trimpath -o bin/webhook ./cmd/webhook
RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -trimpath -o bin/api-server ./cmd/api-server

# Controller image
FROM alpine:3.18 AS controller
RUN apk add --no-cache ca-certificates
WORKDIR /app

# Create a writable directory for the controller binary
# This allows Tilt live_update to work with non-root user (65532)
RUN mkdir -p /app && chmod 777 /app

COPY --from=builder /workspace/bin/controller /app/controller
RUN chmod 755 /app/controller

USER 65532:65532

ENTRYPOINT ["/app/controller"]

# Webhook image
FROM alpine:3.18 AS webhook
RUN apk add --no-cache ca-certificates
WORKDIR /app

# Create a writable directory for the webhook binary
# This allows Tilt live_update to work with non-root user (65532)
RUN mkdir -p /app && chmod 777 /app

COPY --from=builder /workspace/bin/webhook /app/webhook
RUN chmod 755 /app/webhook

USER 65532:65532

ENTRYPOINT ["/app/webhook"]

# API Server image
# Supports both production builds (via Dockerfile) and Tilt development (via docker_build_with_restart)
FROM alpine:3.18 AS api-server
RUN apk add --no-cache ca-certificates
WORKDIR /app

# For production builds: Copy the precompiled binary from builder stage
COPY --from=builder /workspace/bin/api-server /app/api-server

# Copy templates and static assets from source
# In Tilt development, these will be overridden by live_update syncs
COPY cmd/api-server/templates ./templates
COPY cmd/api-server/static ./static

# Create directories with proper permissions for Tilt live_update
# Tilt will sync the binary and assets into these locations during development
RUN mkdir -p /app/templates /app/static && chmod 777 /app /app/templates /app/static

EXPOSE 8080

ENTRYPOINT ["/app/api-server"]
CMD ["-base-dir", "/app"]
