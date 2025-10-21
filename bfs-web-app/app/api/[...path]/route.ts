/* eslint-disable @typescript-eslint/no-explicit-any */
import { cookies, headers } from "next/headers"
import { NextResponse } from "next/server"

// Backend base the server can reach; resolved at runtime in container
function backendBase(): string {
  return process.env.BACKEND_INTERNAL_URL || process.env.INTERNAL_API_BASE_URL || "http://backend:8080"
}

async function handle(req: Request, params: Promise<{ path: string[] }>) {
  const { path } = await params
  const target = new URL(path.join("/"), backendBase())

  const method = req.method
  const inHeaders = new Headers(req.headers)
  const outHeaders = new Headers()

  // Forward Authorization if present
  const auth = inHeaders.get("authorization")
  if (auth) outHeaders.set("authorization", auth)

  // CSRF: validate header matches cookie; then forward header + only csrf cookie
  const headerToken = inHeaders.get("x-csrf") || inHeaders.get("X-CSRF") || undefined
  const hdrs = await headers()
  const proto = (hdrs.get("x-forwarded-proto") || "").toLowerCase()
  const rtCookieName = (proto === "https" ? "__Host-" : "") + "csrf"
  const ck = await cookies()
  const cookieToken = ck.get(rtCookieName)?.value
  if ((method !== "GET" && method !== "HEAD" && method !== "OPTIONS") as boolean) {
    if (!headerToken || !cookieToken || headerToken !== cookieToken) {
      return NextResponse.json({ error: true, message: "Invalid CSRF token" }, { status: 403 })
    }
    outHeaders.set("X-CSRF", headerToken)
    outHeaders.set("cookie", `csrf=${cookieToken}; __Host-csrf=${cookieToken}`)
  }

  // Copy content-type for body forwarding
  const ctype = inHeaders.get("content-type")
  if (ctype) outHeaders.set("content-type", ctype)

  // Build body if present
  let body: BodyInit | undefined
  if (method !== "GET" && method !== "HEAD") {
    const ab = await req.arrayBuffer()
    if (ab && ab.byteLength > 0) {
      body = Buffer.from(ab)
    }
  }

  const res = await fetch(target.toString(), { method, headers: outHeaders, body, redirect: "manual" })

  const respHeaders = new Headers(res.headers)
  // Strip Set-Cookie from backend to avoid third-party cookie issues
  respHeaders.delete("set-cookie")

  const buf = Buffer.from(await res.arrayBuffer())
  return new NextResponse(buf, { status: res.status, headers: respHeaders })
}

export const dynamic = "force-dynamic"

export async function GET(req: Request, ctx: any) {
  return handle(req, ctx.params)
}
export async function POST(req: Request, ctx: any) {
  return handle(req, ctx.params)
}
export async function PUT(req: Request, ctx: any) {
  return handle(req, ctx.params)
}
export async function PATCH(req: Request, ctx: any) {
  return handle(req, ctx.params)
}
export async function DELETE(req: Request, ctx: any) {
  return handle(req, ctx.params)
}
export async function OPTIONS(req: Request, ctx: any) {
  return handle(req, ctx.params)
}
export async function HEAD(req: Request, ctx: any) {
  return handle(req, ctx.params)
}
