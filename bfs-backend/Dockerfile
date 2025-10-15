# syntax=docker/dockerfile:1

# Build stage with security hardening
FROM golang:1.25-alpine AS build

# Install security updates and build dependencies
RUN apk update && \
    apk upgrade && \
    apk add --no-cache ca-certificates git && \
    rm -rf /var/cache/apk/*

# Create non-root user for build
RUN addgroup -g 1001 -S appbuild && \
    adduser -u 1001 -S appbuild -G appbuild

WORKDIR /src

# Copy dependency files first for better layer caching
COPY go.mod go.sum ./

# Ensure build directory is writable by non-root user
RUN chown -R appbuild:appbuild /src

# Download dependencies as non-root user
USER appbuild
RUN go mod download && go mod verify

# Copy source code
COPY --chown=appbuild:appbuild . .

# Build with security hardening flags
ARG BUILD_TAGS=""
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -tags="${BUILD_TAGS}" \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o backend ./cmd/backend

# Runtime stage - minimal distroless image
FROM gcr.io/distroless/static-debian12:nonroot

# Copy CA certificates for HTTPS requests
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy binary with proper permissions
COPY --from=build --chown=nonroot:nonroot /src/backend /app/backend

# Set security environment variables
ENV DOCKER_CONTAINER=true \
    GO_ENV=production \
    GIN_MODE=release

# Use non-root user (uid 65532)
USER nonroot:nonroot

# Set working directory
WORKDIR /app

# Health check for container orchestration
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD ["/app/backend", "--healthcheck"]

# Expose port (non-privileged)
EXPOSE 8080

# Run the binary
ENTRYPOINT ["/app/backend"]
