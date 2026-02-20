import withBundleAnalyzer from "@next/bundle-analyzer"
import { withSentryConfig } from "@sentry/nextjs"
import { type NextConfig } from "next"
import { readFileSync } from "fs"
import { resolve } from "path"

if (!process.env.APP_VERSION) {
  try {
    process.env.APP_VERSION = readFileSync(resolve(process.cwd(), "..", "VERSION"), "utf-8").trim()
  } catch {}
}

if (!process.env.NEXT_PUBLIC_APP_VERSION) {
  process.env.NEXT_PUBLIC_APP_VERSION = process.env.APP_VERSION
}

const config: NextConfig = {
  devIndicators: false,
  reactStrictMode: true,
  output: "standalone",
  logging: {
    fetches: {
      fullUrl: true,
    },
  },
  images: {
    formats: ["image/avif", "image/webp"],
    minimumCacheTTL: 2592000,
    remotePatterns: [
      { protocol: "http", hostname: "localhost", port: "8080", pathname: "/**" },
      { protocol: "http", hostname: "127.0.0.1", port: "8080", pathname: "/**" },
      { protocol: "https", hostname: "*.blessthun.ch", pathname: "/**" },
      { protocol: "https", hostname: "*.food.blessthun.ch", pathname: "/**" },
      { protocol: "https", hostname: "*.blob.core.windows.net", pathname: "/**" },
      { protocol: "http", hostname: "localhost", port: "10000", pathname: "/**" },
      { protocol: "http", hostname: "127.0.0.1", port: "10000", pathname: "/**" },
    ],
    qualities: [90],
  },
}

const sentryWrapped = withSentryConfig(config, {
  sourcemaps: {
    disable: !process.env.CI,
  },
  silent: !process.env.CI,
  telemetry: false,
})

const analyze = process.env.ANALYZE === "true"
export default analyze ? withBundleAnalyzer({ enabled: analyze })(sentryWrapped) : sentryWrapped
