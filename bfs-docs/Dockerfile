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

COPY --chown=nodejs:nodejs package.json pnpm-lock.yaml* yarn.lock* package-lock.json* next.config.* source.config.* ./

RUN --mount=type=cache,target=/home/nodejs/.npm,uid=1001,gid=1001 \
    --mount=type=cache,target=/home/nodejs/.pnpm-store,uid=1001,gid=1001 \
    if [ -f pnpm-lock.yaml ]; then pnpm install --frozen-lockfile; \
    elif [ -f yarn.lock ]; then yarn install --frozen-lockfile; \
    else npm ci; fi

COPY --chown=nodejs:nodejs . .

ENV NODE_ENV=production \
    NEXT_TELEMETRY_DISABLED=1

RUN --mount=type=cache,target=/home/nodejs/.next/cache,uid=1001,gid=1001 \
    if [ -f pnpm-lock.yaml ]; then pnpm run build; \
    elif [ -f yarn.lock ]; then yarn run build; \
    else npm run build; fi

FROM gcr.io/distroless/nodejs24-debian12:nonroot AS runtime

ENV NODE_ENV=production \
    NODE_OPTIONS="--max-old-space-size=512 --no-deprecation --trace-warnings --trace-uncaught --unhandled-rejections=strict" \
    NEXT_TELEMETRY_DISABLED=1 \
    PORT=3000

WORKDIR /app

COPY --from=builder --chown=nonroot:nonroot /app/.next/standalone ./
COPY --from=builder --chown=nonroot:nonroot /app/.next/static ./.next/static

HEALTHCHECK --interval=30s --timeout=10s --start-period=20s --retries=3 \
    CMD ["node", "--eval", "require('http').get('http://localhost:3000/api/health', (res) => process.exit(res.statusCode === 200 ? 0 : 1)).on('error', () => process.exit(1))"]

USER nonroot:nonroot

EXPOSE 3000

CMD ["server.js"]
