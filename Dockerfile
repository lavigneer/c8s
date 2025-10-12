# Multi-stage Dockerfile for C8S components
# Builds controller, api-server, and webhook binaries

# Build stage
FROM golang:1.25-alpine AS builder

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

# Build all binaries
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o bin/controller ./cmd/controller
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o bin/api-server ./cmd/api-server
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o bin/webhook ./cmd/webhook

# Controller image
FROM gcr.io/distroless/static:nonroot AS controller
WORKDIR /
COPY --from=builder /workspace/bin/controller /controller
USER 65532:65532

ENTRYPOINT ["/controller"]

# API Server image
FROM gcr.io/distroless/static:nonroot AS api-server
WORKDIR /
COPY --from=builder /workspace/bin/api-server /api-server
# Copy web assets for optional dashboard
COPY web/ /web/
USER 65532:65532

ENTRYPOINT ["/api-server"]

# Webhook image
FROM gcr.io/distroless/static:nonroot AS webhook
WORKDIR /
COPY --from=builder /workspace/bin/webhook /webhook
USER 65532:65532

ENTRYPOINT ["/webhook"]
