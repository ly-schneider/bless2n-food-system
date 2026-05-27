import * as Sentry from "@sentry/nextjs"
import { cookies } from "next/headers"
import { NextResponse } from "next/server"
import { resolveCookieNames } from "@/lib/server/cookies"

function backendBase(): string {
  return process.env.BACKEND_INTERNAL_URL || "http://backend:8080"
}

// Azure Container Apps scales the backend to zero when idle (see GH issue #340).
// The first request after a cold start surfaces here as either a connect-time
// `TypeError: fetch failed` or an in-flight `Error: failed to pipe response`.
// Both shapes are folded into one Sentry group so they don't keep re-opening
// new issues. See GH issue #343.
const COLD_START_FINGERPRINT = ["backend-proxy-cold-start"]

function isColdStartFetchError(err: unknown): boolean {
  return err instanceof TypeError && /fetch failed/i.test(err.message)
}

function isColdStartPipeError(err: unknown): boolean {
  return err instanceof Error && /failed to pipe response/i.test(err.message)
}

function reportColdStart(err: unknown, kind: "fetch_failed" | "pipe_failed") {
  Sentry.captureException(err, {
    fingerprint: COLD_START_FINGERPRINT,
    tags: { cause: "cold_start", proxy_failure: kind },
  })
}

async function handle(req: Request, params: Promise<{ path: string[] }>) {
  const { path } = await params
  const target = new URL(path.join("/"), backendBase())

  // Forward query parameters from the original request
  const reqUrl = new URL(req.url)
  target.search = reqUrl.search

  const method = req.method
  const inHeaders = new Headers(req.headers)
  const outHeaders = new Headers()

  const auth = inHeaders.get("authorization")
  if (auth) outHeaders.set("authorization", auth)

  const contentType = inHeaders.get("content-type")
  if (contentType) outHeaders.set("content-type", contentType)

  const cookieIn = inHeaders.get("cookie")
  if (cookieIn) outHeaders.set("cookie", cookieIn)

  const idempotencyKey = inHeaders.get("idempotency-key") || inHeaders.get("Idempotency-Key")
  if (idempotencyKey) outHeaders.set("Idempotency-Key", idempotencyKey)

  const headerToken = inHeaders.get("x-csrf") || inHeaders.get("X-CSRF") || undefined
  const { csrfName } = await resolveCookieNames()
  const ck = await cookies()
  const cookieToken = ck.get(csrfName)?.value
  const isMutating = !(method === "GET" || method === "HEAD" || method === "OPTIONS")

  const pathName = target.pathname
  const csrfExempt =
    // Public or device-gated endpoints
    pathName.startsWith("/v1/pos/") ||
    pathName.startsWith("/v1/stations/") ||
    pathName.startsWith("/v1/devices/") ||
    pathName.startsWith("/v1/health") ||
    pathName.startsWith("/v1/public/") ||
    pathName.startsWith("/v1/invites/") ||
    pathName.startsWith("/v1/payments/webhooks/") ||
    pathName.startsWith("/v1/claim/")

  const hasBearerAuth = !!inHeaders.get("authorization")?.startsWith("Bearer ")

  if (isMutating && !csrfExempt && !hasBearerAuth) {
    if (!headerToken || !cookieToken || headerToken !== cookieToken) {
      return NextResponse.json({ error: true, message: "Invalid CSRF token" }, { status: 403 })
    }
  }
  // Always forward the CSRF header when present so the backend's own
  // CSRF middleware can validate it (e.g. admin-gated device pairing).
  if (headerToken) {
    outHeaders.set("X-CSRF", headerToken)
  }

  let body: BodyInit | undefined
  if (method !== "GET" && method !== "HEAD") {
    const ab = await req.arrayBuffer()
    if (ab && ab.byteLength > 0) {
      body = Buffer.from(ab)
    }
  }

  let res: Response
  try {
    res = await fetch(target.toString(), { method, headers: outHeaders, body, redirect: "manual" })
  } catch (err) {
    if (isColdStartFetchError(err)) {
      reportColdStart(err, "fetch_failed")
      return NextResponse.json(
        { error: true, message: "Backend temporarily unavailable", cause: "cold_start" },
        { status: 502 }
      )
    }
    throw err
  }

  const respHeaders = new Headers(res.headers)

  if (res.status === 204 || res.status === 205 || res.status === 304) {
    return new NextResponse(null, { status: res.status, headers: respHeaders })
  }

  // Stream SSE responses instead of buffering
  const respContentType = res.headers.get("content-type") || ""
  if (respContentType.includes("text/event-stream") && res.body) {
    return new NextResponse(res.body, { status: res.status, headers: respHeaders })
  }

  try {
    const buf = Buffer.from(await res.arrayBuffer())
    return new NextResponse(buf, { status: res.status, headers: respHeaders })
  } catch (err) {
    if (isColdStartFetchError(err) || isColdStartPipeError(err)) {
      reportColdStart(err, "pipe_failed")
      return NextResponse.json(
        { error: true, message: "Backend response interrupted", cause: "cold_start" },
        { status: 504 }
      )
    }
    throw err
  }
}

export const dynamic = "force-dynamic"

export async function GET(req: Request, ctx: { params: Promise<{ path: string[] }> }) {
  return handle(req, ctx.params)
}
export async function POST(req: Request, ctx: { params: Promise<{ path: string[] }> }) {
  return handle(req, ctx.params)
}
export async function PUT(req: Request, ctx: { params: Promise<{ path: string[] }> }) {
  return handle(req, ctx.params)
}
export async function PATCH(req: Request, ctx: { params: Promise<{ path: string[] }> }) {
  return handle(req, ctx.params)
}
export async function DELETE(req: Request, ctx: { params: Promise<{ path: string[] }> }) {
  return handle(req, ctx.params)
}
export async function OPTIONS(req: Request, ctx: { params: Promise<{ path: string[] }> }) {
  return handle(req, ctx.params)
}
export async function HEAD(req: Request, ctx: { params: Promise<{ path: string[] }> }) {
  return handle(req, ctx.params)
}
