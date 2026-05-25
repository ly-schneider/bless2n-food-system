import * as Sentry from "@sentry/nextjs"

export async function register() {
  if (process.env.NEXT_RUNTIME === "nodejs") {
    await import("./sentry.server.config")
  }
  if (process.env.NEXT_RUNTIME === "edge") {
    await import("./sentry.edge.config")
  }
}

const COLD_START_FINGERPRINT = ["backend-proxy-cold-start"]

function isColdStartProxyError(err: unknown, path: string | undefined): boolean {
  if (!path?.startsWith("/api/") || path.startsWith("/api/auth/")) return false
  if (!(err instanceof Error)) return false
  return /fetch failed/i.test(err.message) || /failed to pipe response/i.test(err.message)
}

export const onRequestError: typeof Sentry.captureRequestError = (err, request, errorContext) => {
  if (isColdStartProxyError(err, request.path)) {
    Sentry.captureException(err, {
      fingerprint: COLD_START_FINGERPRINT,
      tags: { cause: "cold_start", proxy_failure: "escaped_request_error" },
    })
    return
  }
  return Sentry.captureRequestError(err, request, errorContext)
}
