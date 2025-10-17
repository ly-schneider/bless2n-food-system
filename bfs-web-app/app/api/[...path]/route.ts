import { NextRequest, NextResponse } from "next/server"

// Backend base the server can reach; resolved at runtime in container
function backendBase(): string {
  return (
    process.env.BACKEND_INTERNAL_URL ||
    process.env.INTERNAL_API_BASE_URL ||
    "http://backend:8080"
  )
}

async function handle(req: NextRequest, params: { path: string[] }) {
  const { path } = params
  const target = new URL(path.join("/"), backendBase())

  const method = req.method
  const inHeaders = new Headers(req.headers)
  const outHeaders = new Headers()

  // Forward Authorization if present
  const auth = inHeaders.get("authorization")
  if (auth) outHeaders.set("authorization", auth)

  // CSRF: validate header matches cookie; then forward header + only csrf cookie
  const headerToken = inHeaders.get("x-csrf") || inHeaders.get("X-CSRF") || undefined
  const cookieToken = req.cookies.get((req.nextUrl.protocol === "https:" ? "__Host-" : "") + "csrf")?.value
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

export async function GET(req: NextRequest, ctx: { params: { path: string[] } }) {
  return handle(req, ctx.params)
}
export async function POST(req: NextRequest, ctx: { params: { path: string[] } }) {
  return handle(req, ctx.params)
}
export async function PUT(req: NextRequest, ctx: { params: { path: string[] } }) {
  return handle(req, ctx.params)
}
export async function PATCH(req: NextRequest, ctx: { params: { path: string[] } }) {
  return handle(req, ctx.params)
}
export async function DELETE(req: NextRequest, ctx: { params: { path: string[] } }) {
  return handle(req, ctx.params)
}
export async function OPTIONS(req: NextRequest, ctx: { params: { path: string[] } }) {
  return handle(req, ctx.params)
}
export async function HEAD(req: NextRequest, ctx: { params: { path: string[] } }) {
  return handle(req, ctx.params)
}
