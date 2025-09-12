import withBundleAnalyzer from "@next/bundle-analyzer"
import { type NextConfig } from "next"

import { env } from "./env.mjs"

const API_BASE = process.env.NEXT_PUBLIC_API_BASE_URL || "http://localhost:8080"

const config: NextConfig = {
  devIndicators: false,
  reactStrictMode: true,
  logging: {
    fetches: {
      fullUrl: true,
    },
  },
  rewrites: async () => [
    { source: "/healthz", destination: `${API_BASE}/ping` },
    { source: "/api/healthz", destination: `${API_BASE}/ping` },
    { source: "/health", destination: `${API_BASE}/ping` },
    { source: "/ping", destination: `${API_BASE}/ping` },
  ],
  eslint: {
    dirs: ['.'],
  },
}

export default env.ANALYZE ? withBundleAnalyzer({ enabled: env.ANALYZE })(config) : config
