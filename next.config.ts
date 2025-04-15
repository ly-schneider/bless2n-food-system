import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  devIndicators: false,
  images: {
    remotePatterns: [
      new URL(
        "https://upload.wikimedia.org/**"
      ),
    ],
  },
};

export default nextConfig;
