import * as Sentry from "@sentry/nextjs"

export async function register() {
  if (process.env.NEXT_RUNTIME === "nodejs") {
    await configureBackendKeepAlive()
    await import("./sentry.server.config")
  }
  if (process.env.NEXT_RUNTIME === "edge") {
    await import("./sentry.edge.config")
  }
}

async function configureBackendKeepAlive() {
  try {
    const { Agent, setGlobalDispatcher } = await import("undici")
    setGlobalDispatcher(
      new Agent({
        keepAliveTimeout: 30_000,
        keepAliveMaxTimeout: 60_000,
        connections: 64,
        pipelining: 1,
        headersTimeout: 0,
        bodyTimeout: 0,
      })
    )
  } catch (err) {
    console.warn("undici keep-alive setup failed:", err)
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
