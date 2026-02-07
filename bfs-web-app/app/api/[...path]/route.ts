import { cookies } from "next/headers"
import { NextResponse } from "next/server"
import { resolveCookieNames } from "@/lib/server/cookies"

function backendBase(): string {
  return process.env.BACKEND_INTERNAL_URL || "http://backend:8080"
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
    pathName.startsWith("/v1/payments/webhooks/")

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

  const res = await fetch(target.toString(), { method, headers: outHeaders, body, redirect: "manual" })

  const respHeaders = new Headers(res.headers)

  if (res.status === 204 || res.status === 205 || res.status === 304) {
    return new NextResponse(null, { status: res.status, headers: respHeaders })
  }

  // Stream SSE responses instead of buffering
  const respContentType = res.headers.get("content-type") || ""
  if (respContentType.includes("text/event-stream") && res.body) {
    return new NextResponse(res.body, { status: res.status, headers: respHeaders })
  }

  const buf = Buffer.from(await res.arrayBuffer())
  return new NextResponse(buf, { status: res.status, headers: respHeaders })
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
