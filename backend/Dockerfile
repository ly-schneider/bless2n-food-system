# syntax=docker/dockerfile:1.6
###############################################################################
# 1. Build stage – always Go
###############################################################################
FROM golang:1.24-alpine AS build
WORKDIR /app

# (cache) download modules
COPY go.mod go.sum ./
RUN go mod download

# Air (tiny ~10 MB) – just install unconditionally
RUN go install github.com/air-verse/air@latest   # live-reload :contentReference[oaicite:2]{index=2}

# compile the service
COPY . .
RUN go build -o /server server.go

###############################################################################
# 2a. Runtime stage for *development*  (has Go + Air)
###############################################################################
FROM golang:1.24-alpine AS development
ARG  APP_ENV=development
ENV  APP_ENV=$APP_ENV
WORKDIR /app
COPY --from=build /server .
COPY --from=build /go/bin/air /usr/local/bin/air
EXPOSE 8080
CMD ["sh","-c","air"]                # always hot-reload in dev

###############################################################################
# 2b. Runtime stage for *production*  (tiny, no Go)
###############################################################################
FROM alpine:3.20 AS production
ARG  APP_ENV=production
ENV  APP_ENV=$APP_ENV
WORKDIR /app
COPY --from=build /server .
EXPOSE 8080
CMD ["./server"]
