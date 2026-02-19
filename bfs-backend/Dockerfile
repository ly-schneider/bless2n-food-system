# syntax=docker/dockerfile:1

# ── Stage 1: Bundle OpenAPI spec ──
FROM node:22-alpine AS openapi
WORKDIR /src
COPY openapi/ openapi/
RUN npx --yes @redocly/cli bundle openapi/openapi.yaml -o openapi/bundled.yaml

# ── Stage 2: Build Go binary ──
FROM golang:1.25-alpine AS build

RUN apk update && \
    apk upgrade && \
    apk add --no-cache ca-certificates git && \
    rm -rf /var/cache/apk/*

RUN addgroup -g 1001 -S appbuild && \
    adduser -u 1001 -S appbuild -G appbuild

WORKDIR /src

COPY go.mod go.sum ./

RUN chown -R appbuild:appbuild /src

USER appbuild
RUN --mount=type=cache,target=/go/pkg/mod,uid=1001,gid=1001 \
    go mod download && go mod verify

USER root
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    go install github.com/swaggo/swag/cmd/swag@v1.16.6 && \
    go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.5.1

USER appbuild
COPY --chown=appbuild:appbuild . .
COPY --from=openapi --chown=appbuild:appbuild /src/openapi/bundled.yaml openapi/bundled.yaml

# Generate ent ORM code
RUN --mount=type=cache,target=/go/pkg/mod,uid=1001,gid=1001 \
    mkdir -p ./internal/generated/ent && \
    echo "package ent" > ./internal/generated/ent/ent.go && \
    go run -mod=mod entgo.io/ent/cmd/ent generate \
        --feature sql/modifier --feature sql/upsert \
        --target ./internal/generated/ent ./internal/schema

# Generate API types from bundled OpenAPI spec
RUN oapi-codegen -config oapi-codegen.yaml openapi/bundled.yaml

# Generate Swagger docs
RUN --mount=type=cache,target=/go/pkg/mod,uid=1001,gid=1001 \
    swag init -g cmd/backend/main.go --parseInternal --parseDependency --generatedTime=false

ARG BUILD_TAGS="docs"
RUN --mount=type=cache,target=/home/appbuild/.cache/go-build,uid=1001,gid=1001 \
    --mount=type=cache,target=/go/pkg/mod,uid=1001,gid=1001 \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -tags="${BUILD_TAGS}" \
    -ldflags="-w -s -extldflags '-static'" \
    -o backend ./cmd/backend

# ── Stage 3: Runtime ──
FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build --chown=nonroot:nonroot /src/backend /app/backend

ENV DOCKER_CONTAINER=true \
    GO_ENV=production \
    GIN_MODE=release

USER nonroot:nonroot
WORKDIR /app

HEALTHCHECK --interval=60s --timeout=5s --start-period=20s --retries=3 \
    CMD ["/app/backend", "--healthcheck"]

EXPOSE 8080

ENTRYPOINT ["/app/backend"]
