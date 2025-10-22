# syntax=docker/dockerfile:1

# ---- 1) Builder: Debian slim with security hardening
FROM node:24-bookworm-slim AS builder

# Install security updates
RUN apt-get update && \
    apt-get upgrade -y && \
    apt-get install -y --no-install-recommends \
        ca-certificates \
        && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

# Create non-root user for build
RUN groupadd --gid 1001 --system nodejs && \
    useradd --uid 1001 --system --gid nodejs --create-home nodejs

# Set working directory and ownership
WORKDIR /app
RUN chown nodejs:nodejs /app

# Enable package manager (requires root to create shims)
RUN corepack enable

# Switch to non-root user
USER nodejs

# Copy package files with proper ownership
COPY --chown=nodejs:nodejs package.json pnpm-lock.yaml* yarn.lock* package-lock.json* ./

# Install dependencies with cache mounts (as non-root)
RUN --mount=type=cache,target=/home/nodejs/.npm,uid=1001,gid=1001 \
    --mount=type=cache,target=/home/nodejs/.pnpm-store,uid=1001,gid=1001 \
    if [ -f pnpm-lock.yaml ]; then pnpm install --frozen-lockfile; \
    elif [ -f yarn.lock ]; then yarn install --frozen-lockfile; \
    else npm ci; fi

# Copy source code
COPY --chown=nodejs:nodejs . .

# Production build flags only; no app runtime env baked here
ENV NODE_ENV=production \
    NEXT_TELEMETRY_DISABLED=1

# Build using the same package manager used for install; skip npm prune to avoid
# unsupported protocols (e.g., link:) and because Next.js standalone bundles deps.
RUN if [ -f pnpm-lock.yaml ]; then pnpm run build; \
    elif [ -f yarn.lock ]; then yarn run build; \
    else npm run build; fi

# ---- 2) Runtime: Distroless with security hardening
FROM gcr.io/distroless/nodejs24-debian12:nonroot AS runtime

# Set secure environment variables (runtime env like BACKEND_INTERNAL_URL and
# NEXT_PUBLIC_STRIPE_PUBLISHABLE_KEY are expected to be injected by the orchestrator at start)
ENV NODE_ENV=production \
    NODE_OPTIONS="--max-old-space-size=2048 --no-deprecation --trace-warnings --trace-uncaught --unhandled-rejections=strict" \
    NEXT_TELEMETRY_DISABLED=1 \
    PORT=3000

# Set working directory
WORKDIR /app

# Copy built application with proper ownership
COPY --from=builder --chown=nonroot:nonroot /app/.next/standalone ./
COPY --from=builder --chown=nonroot:nonroot /app/.next/static ./.next/static
COPY --from=builder --chown=nonroot:nonroot /app/public ./public

# Health check for container orchestration
HEALTHCHECK --interval=30s --timeout=10s --start-period=20s --retries=3 \
    CMD ["node", "--eval", "require('http').get('http://localhost:3000/api/health', (res) => process.exit(res.statusCode === 200 ? 0 : 1)).on('error', () => process.exit(1))"]

# Use non-root user (uid 65532 - nonroot)
USER nonroot:nonroot

# Expose port (non-privileged)
EXPOSE 3000

# Start the application
CMD ["server.js"]
