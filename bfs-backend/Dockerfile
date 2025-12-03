# syntax=docker/dockerfile:1

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
    go install github.com/swaggo/swag/cmd/swag@v1.16.6

USER appbuild
COPY --chown=appbuild:appbuild . .

RUN --mount=type=cache,target=/go/pkg/mod,uid=1001,gid=1001 \
    swag init -g cmd/backend/main.go --parseInternal --parseDependency --generatedTime=false

ARG BUILD_TAGS="docs"
RUN --mount=type=cache,target=/home/appbuild/.cache/go-build,uid=1001,gid=1001 \
    --mount=type=cache,target=/go/pkg/mod,uid=1001,gid=1001 \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -tags="${BUILD_TAGS}" \
    -ldflags='-w -s -extldflags "-static"' \
    -o backend ./cmd/backend

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
