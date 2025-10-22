import { cookies } from "next/headers"
import { NextResponse } from "next/server"
import { resolveCookieNames } from "@/lib/server/cookies"

function backendBase(): string {
  return process.env.BACKEND_INTERNAL_URL || process.env.INTERNAL_API_BASE_URL || "http://backend:8080"
}

async function handle(req: Request, params: Promise<{ path: string[] }>) {
  const { path } = await params
  const target = new URL(path.join("/"), backendBase())

  const method = req.method
  const inHeaders = new Headers(req.headers)
  const outHeaders = new Headers()

  const auth = inHeaders.get("authorization")
  if (auth) outHeaders.set("authorization", auth)
  const contentType = inHeaders.get("content-type")
  if (contentType) outHeaders.set("content-type", contentType)
  const cookieIn = inHeaders.get("cookie")
  if (cookieIn) outHeaders.set("cookie", cookieIn)
  // Forward custom headers needed by backend endpoints
  const xPosToken = inHeaders.get("x-pos-token") || inHeaders.get("X-Pos-Token")
  if (xPosToken) outHeaders.set("X-Pos-Token", xPosToken)
  const xStationKey = inHeaders.get("x-station-key") || inHeaders.get("X-Station-Key")
  if (xStationKey) outHeaders.set("X-Station-Key", xStationKey)
  const idempotencyKey = inHeaders.get("idempotency-key") || inHeaders.get("Idempotency-Key")
  if (idempotencyKey) outHeaders.set("Idempotency-Key", idempotencyKey)

  const headerToken = inHeaders.get("x-csrf") || inHeaders.get("X-CSRF") || undefined
  const { csrfName } = await resolveCookieNames()
  const ck = await cookies()
  const cookieToken = ck.get(csrfName)?.value
  const isMutating = !(method === "GET" || method === "HEAD" || method === "OPTIONS")

  const pathName = target.pathname
  const csrfExempt =
    pathName.startsWith("/v1/pos/") ||
    pathName.startsWith("/v1/stations/") ||
    pathName.startsWith("/v1/health") ||
    pathName.startsWith("/v1/public/")

  if (isMutating && !csrfExempt) {
    if (!headerToken || !cookieToken || headerToken !== cookieToken) {
      return NextResponse.json({ error: true, message: "Invalid CSRF token" }, { status: 403 })
    }
    outHeaders.set("X-CSRF", headerToken)
  }

  let body: BodyInit | undefined
  if (method !== "GET" && method !== "HEAD") {
    const ab = await req.arrayBuffer()
    if (ab && ab.byteLength > 0) {
      body = Buffer.from(ab)
    }
  }

  const start = Date.now()
  let res: Response
  try {
    res = await fetch(target.toString(), { method, headers: outHeaders, body, redirect: "manual" })
  } finally {
    const dur = Date.now() - start
    try {
      const ua = inHeaders.get("user-agent") || "-"
      const fwd = inHeaders.get("x-forwarded-for") || "-"
      console.log(
        `[api-proxy] ${method} ${target.pathname} -> ${target.origin} dur=${dur}ms ip=${
          fwd.split(",")[0]?.trim() || "-"
        } ua=${ua.substring(0, 80)}`
      )
    } catch {}
  }

  const respHeaders = new Headers(res.headers)

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
