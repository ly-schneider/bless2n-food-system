import { createMDX } from "fumadocs-mdx/next";

if (!process.env.APP_VERSION) {
  process.env.APP_VERSION = "dev";
}

const withMDX = createMDX();

/** @type {import('next').NextConfig} */
const config = {
  output: "standalone",
  reactStrictMode: true,
  devIndicators: false,
};

export default withMDX(config);
