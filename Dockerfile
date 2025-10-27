# Multi-stage Dockerfile for C8S components
# Builds controller, api-server, and webhook binaries

# Build stage
FROM golang:1.24-alpine AS builder

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

# Build deployed components for local development
# Use -trimpath to remove build paths, -v for verbose output
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -trimpath -v -o bin/controller ./cmd/controller
# NOTE: api-server build skipped (not deployed in dev environment)
# RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o bin/api-server ./cmd/api-server
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -trimpath -v -o bin/webhook ./cmd/webhook

# Controller image
FROM alpine:3.18 AS controller
RUN apk add --no-cache ca-certificates
WORKDIR /
COPY --from=builder /workspace/bin/controller /controller
USER 65532:65532

ENTRYPOINT ["/controller"]

# API Server image
FROM alpine:3.18 AS api-server
RUN apk add --no-cache ca-certificates
WORKDIR /
COPY --from=builder /workspace/bin/api-server /api-server
# Copy web assets for optional dashboard
COPY web/ /web/
USER 65532:65532

ENTRYPOINT ["/api-server"]

# Webhook image
FROM alpine:3.18 AS webhook
RUN apk add --no-cache ca-certificates
WORKDIR /
COPY --from=builder /workspace/bin/webhook /webhook
USER 65532:65532

ENTRYPOINT ["/webhook"]
