# ---- 1) Builder: Debian slim for compatibility
FROM node:24-bookworm-slim AS builder
ENV NODE_ENV=production
WORKDIR /app

RUN corepack enable

COPY package.json pnpm-lock.yaml* yarn.lock* package-lock.json* ./
RUN --mount=type=cache,target=/root/.npm \
    --mount=type=cache,target=/root/.pnpm-store \
    if [ -f pnpm-lock.yaml ]; then pnpm install --frozen-lockfile; \
    elif [ -f yarn.lock ]; then yarn install --frozen-lockfile; \
    else npm ci; fi

COPY . .

RUN npm run build && npm prune --omit=dev

# ---- 2) Runtime: Distroless (max security)
FROM gcr.io/distroless/nodejs24-debian12 AS runtime
ENV NODE_ENV=production
WORKDIR /app

COPY --from=builder /app/.next/standalone ./
COPY --from=builder /app/.next/static ./.next/static
COPY --from=builder /app/public ./public

EXPOSE 3000
CMD ["server.js"]  # or the entry point Next.js generated in standalone