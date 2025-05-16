export default {
  async fetch(request, env, ctx) {
    const url = new URL(request.url)
    if (url.pathname.startsWith('/api/')) {
      // Handle API routes
      return env.ASSETS.fetch(request)
    }
    // Serve static assets
    return env.ASSETS.fetch(request)
  },
}

export const config = {
  compatibility_flags: ["nodejs_compat"],
}
