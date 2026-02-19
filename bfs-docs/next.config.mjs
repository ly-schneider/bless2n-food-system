import { readFileSync } from 'fs';
import { createMDX } from 'fumadocs-mdx/next';
import { resolve } from 'path';

if (!process.env.APP_VERSION) {
  try {
    process.env.APP_VERSION = readFileSync(resolve(process.cwd(), '..', 'VERSION'), 'utf-8').trim();
  } catch {}
}

const withMDX = createMDX();

/** @type {import('next').NextConfig} */
const config = {
  output: 'standalone',
  reactStrictMode: true,
  devIndicators: false
};

export default withMDX(config);
