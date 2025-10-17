import { NextRequest, NextResponse } from "next/server"

// Backend base the server can reach; fall back to localhost in dev
const BACKEND_BASE =
  process.env.BACKEND_INTERNAL_URL || process.env.INTERNAL_API_BASE_URL || "http://backend:8080"

export async function POST(req: NextRequest) {
  try {
    const { url, method, headers, body } = (await req.json()) as {
      url: string
      method: string
      headers?: Record<string, string>
      body?: string | null
    }

    // Basic CSRF double-submit check
    const headerToken = req.headers.get("x-csrf") || req.headers.get("X-CSRF")
    const cookieToken = req.cookies.get((req.nextUrl.protocol === "https:" ? "__Host-" : "") + "csrf")?.value
    if (!headerToken || !cookieToken || headerToken !== cookieToken) {
      return NextResponse.json({ error: true, message: "Invalid CSRF token" }, { status: 403 })
    }

    // Normalize target URL: support absolute backend URL or relative /api
    let target: URL
    if (url.startsWith("/api/")) {
      target = new URL(url.replace(/^\/api\//, "/"), BACKEND_BASE)
    } else {
      target = new URL(url)
      const backendHost = new URL(BACKEND_BASE)
      if (target.host !== backendHost.host) {
        return NextResponse.json({ error: true, message: "Forbidden target host" }, { status: 400 })
      }
    }

    const forwardHeaders = new Headers(headers || {})
    // Ensure backend CSRF middleware receives valid tokens
    // 1) Forward the validated CSRF header
    if (headerToken) {
      forwardHeaders.set("X-CSRF", headerToken)
    }
    // 2) Forward only the CSRF cookie (not all client cookies)
    //    Backend accepts either name; send both to cover http/https configs.
    forwardHeaders.set("cookie", `csrf=${cookieToken}; __Host-csrf=${cookieToken}`)

    const res = await fetch(target.toString(), {
      method,
      headers: forwardHeaders,
      body: body ?? undefined,
      redirect: "manual",
    })

    const outHeaders = new Headers(res.headers)
    // Strip Set-Cookie to avoid cross-site cookie issues in WebView
    outHeaders.delete("set-cookie")

    const buffer = Buffer.from(await res.arrayBuffer())
    return new NextResponse(buffer, {
      status: res.status,
      headers: outHeaders,
    })
  } catch (e: unknown) {
    const msg = e instanceof Error ? e.message : "Proxy error"
    return NextResponse.json({ error: true, message: msg }, { status: 500 })
  }
}
