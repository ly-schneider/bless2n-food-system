# syntax=docker/dockerfile:1

FROM node:24-bookworm-slim AS builder

RUN apt-get update && \
    apt-get upgrade -y && \
    apt-get install -y --no-install-recommends ca-certificates && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

RUN groupadd --gid 1001 --system nodejs && \
    useradd --uid 1001 --system --gid nodejs --create-home nodejs

WORKDIR /app
RUN chown nodejs:nodejs /app

RUN corepack enable

USER nodejs

COPY --chown=nodejs:nodejs package.json pnpm-lock.yaml* yarn.lock* package-lock.json* ./

RUN --mount=type=cache,target=/home/nodejs/.npm,uid=1001,gid=1001 \
    --mount=type=cache,target=/home/nodejs/.pnpm-store,uid=1001,gid=1001 \
    if [ -f pnpm-lock.yaml ]; then pnpm install --frozen-lockfile; \
    elif [ -f yarn.lock ]; then yarn install --frozen-lockfile; \
    else npm ci; fi

COPY --chown=nodejs:nodejs . .

ARG SENTRY_AUTH_TOKEN
ARG SENTRY_ORG
ARG SENTRY_PROJECT
ARG CI

ENV NODE_ENV=production \
    NEXT_TELEMETRY_DISABLED=1

RUN --mount=type=cache,target=/home/nodejs/.next/cache,uid=1001,gid=1001 \
    if [ -f pnpm-lock.yaml ]; then pnpm run build; \
    elif [ -f yarn.lock ]; then yarn run build; \
    else npm run build; fi

FROM node:24-alpine AS runtime

RUN addgroup --system --gid 1001 nodejs && \
    adduser --system --uid 1001 nodejs

ENV NODE_ENV=production \
    NODE_OPTIONS="--max-old-space-size=384 --no-deprecation" \
    NEXT_TELEMETRY_DISABLED=1 \
    HOSTNAME=0.0.0.0 \
    PORT=3000

WORKDIR /app

COPY --from=builder --chown=nodejs:nodejs /app/.next/standalone ./
COPY --from=builder --chown=nodejs:nodejs /app/.next/static ./.next/static
COPY --from=builder --chown=nodejs:nodejs /app/public ./public

USER nodejs
EXPOSE 3000
CMD ["node", "server.js"]
