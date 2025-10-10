import withBundleAnalyzer from "@next/bundle-analyzer"
import { type NextConfig } from "next"

const API_BASE = process.env.BACKEND_INTERNAL_URL || process.env.NEXT_PUBLIC_API_BASE_URL || "http://backend:8080"

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
    remotePatterns: [
      { protocol: "http", hostname: "localhost", port: "8080", pathname: "/**" },
      { protocol: "http", hostname: "127.0.0.1", port: "8080", pathname: "/**" },
      // Allow any https source (CDNs, object storage) used by the API
      { protocol: "https", hostname: "**", pathname: "/**" },
    ],
  },
  rewrites: async () => [
    { source: "/healthz", destination: `${API_BASE}/ping` },
    { source: "/api/healthz", destination: `${API_BASE}/ping` },
    { source: "/health", destination: `${API_BASE}/ping` },
    { source: "/ping", destination: `${API_BASE}/ping` },
  ],
  eslint: {
    dirs: ["."],
  },
}

const analyze = process.env.ANALYZE === "true"
export default analyze ? withBundleAnalyzer({ enabled: analyze })(config) : config
