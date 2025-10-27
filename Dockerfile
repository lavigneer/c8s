# Multi-stage Dockerfile for C8S components
# Builds controller and webhook binaries

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

# Build deployed components
# Use -trimpath to remove build paths, -v for verbose output
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -trimpath -v -o bin/controller ./cmd/controller
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -trimpath -v -o bin/webhook ./cmd/webhook
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -trimpath -v -o bin/api-server ./cmd/api-server

# Controller image
FROM alpine:3.18 AS controller
RUN apk add --no-cache ca-certificates
WORKDIR /
COPY --from=builder /workspace/bin/controller /controller
USER 65532:65532

ENTRYPOINT ["/controller"]

# Webhook image
FROM alpine:3.18 AS webhook
RUN apk add --no-cache ca-certificates
WORKDIR /
COPY --from=builder /workspace/bin/webhook /webhook
USER 65532:65532

ENTRYPOINT ["/webhook"]

# API Server image
FROM alpine:3.18 AS api-server
RUN apk add --no-cache ca-certificates
WORKDIR /app
# Copy the api-server binary
COPY --from=builder /workspace/bin/api-server /app/api-server
# Copy templates and static assets
COPY cmd/api-server/templates ./templates
COPY cmd/api-server/static ./static
EXPOSE 8080
USER 65532:65532

ENTRYPOINT ["/app/api-server"]
CMD ["-base-dir", "/app"]
