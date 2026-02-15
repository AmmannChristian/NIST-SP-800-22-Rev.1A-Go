# Multi-stage build for the NIST SP 800-22 service using the pure Go implementation.

# Builder: compile the Go service
FROM golang:1.25-alpine AS build

WORKDIR /app
RUN apk add --no-cache git make
ENV GO111MODULE=on

# Go module files first for better layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the remaining source
COPY . .

# Build static binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/bin/nist-sp800-22-rev1a ./cmd/server

# Runtime image
FROM alpine:3.18

LABEL maintainer="Christian Ammann" \
      description="NIST SP 800-22 Statistical Test Suite Microservice"

WORKDIR /app

# Minimal runtime deps and non-root user
RUN apk add --no-cache ca-certificates libc6-compat tzdata && \
    addgroup -g 1000 nist && \
    adduser -D -u 1000 -G nist nist

# Copy binary
COPY --from=build /app/bin/nist-sp800-22-rev1a /usr/local/bin/nist-sp800-22-rev1a

# Environment defaults
ENV GRPC_PORT=9090 \
    METRICS_PORT=9091 \
    LOG_LEVEL=info

# Expose ports
EXPOSE 9090 9091

USER nist

ENTRYPOINT ["/usr/local/bin/nist-sp800-22-rev1a"]
