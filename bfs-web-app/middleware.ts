import { NextRequest, NextResponse } from "next/server"

export async function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl

  // Skip static assets
  const isStatic = pathname.startsWith("/_next/") || pathname.includes(".")
  if (!isStatic) {
    try {
      const method = request.method
      const ua = request.headers.get("user-agent") || "-"
      const fwd = request.headers.get("x-forwarded-for") || "-"
      const proto = (request.headers.get("x-forwarded-proto") || "").toLowerCase()
      // Lightweight request log (stdout)
      console.log(`[req] ${method} ${pathname} proto=${proto || "-"} ip=${fwd.split(",")[0]?.trim() || "-"} ua=${ua.substring(0, 80)}`)
    } catch {}
  }

  if (isStatic || pathname.startsWith("/api/")) {
    return NextResponse.next()
  }

  // Auth guard: protect app routes by requiring refresh cookie
  // Note: Checkout and Orders are intentionally NOT protected to allow guest flows
  const protectedPrefixes = ["/profile", "/admin"]
  const isProtected = protectedPrefixes.some((p) => pathname.startsWith(p))
  if (isProtected) {
    const proto = (request.headers.get("x-forwarded-proto") || "").toLowerCase()
    const rtName = proto === "https" ? "__Host-rt" : "rt"
    const hasRefresh = request.cookies.get(rtName)?.value
    if (!hasRefresh) {
      const url = request.nextUrl.clone()
      url.pathname = "/login"
      url.searchParams.set("next", request.nextUrl.pathname)
      return NextResponse.redirect(url)
    }
  }

  // Set security headers for all responses
  const response = NextResponse.next()

  // Security headers
  response.headers.set("X-Frame-Options", "DENY")
  response.headers.set("X-Content-Type-Options", "nosniff")
  response.headers.set("Referrer-Policy", "origin-when-cross-origin")

  // CSP optimized for WebView; API calls go through same-origin /api via Next route
  const connectSrc = ["'self'", "ws:", "wss:"]

  const stripeScript = "https://js.stripe.com"
  const stripeConnect = ["https://api.stripe.com", "https://m.stripe.network", "https://r.stripe.com"]
  const stripeFrame = ["https://js.stripe.com", "https://hooks.stripe.com", "https://m.stripe.network"]

  const googleTagManager = "https://www.googletagmanager.com"

  const csp = [
    "default-src 'self'",
    // Load Stripe.js from Stripe CDN
    `script-src 'self' 'unsafe-inline' 'unsafe-eval' ${stripeScript} ${googleTagManager}`,
    "style-src 'self' 'unsafe-inline'",
    // Stripe Elements iframes and assets
    "img-src 'self' data: blob: https:",
    "font-src 'self' data:",
    `connect-src ${[...connectSrc, ...stripeConnect].join(" ")}`,
    `frame-src ${stripeFrame.join(" ")}`,
    "object-src 'none'",
    "base-uri 'self'",
  ].join("; ")

  response.headers.set("Content-Security-Policy", csp)

  // Ensure a CSRF cookie exists for double-submit pattern (for guests too)
  try {
    const proto = (request.headers.get("x-forwarded-proto") || "").toLowerCase()
    const secure = proto === "https"
    const csrfName = secure ? "__Host-csrf" : "csrf"
    const hasCsrf = request.cookies.get(csrfName)?.value
    if (!hasCsrf) {
      const bytes = crypto.getRandomValues(new Uint8Array(16))
      const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_"
      let token = ""
      for (let i = 0; i < bytes.length; i++) token += chars[bytes[i]! % chars.length]
      response.cookies.set({
        name: csrfName,
        value: token,
        httpOnly: false,
        secure,
        sameSite: "lax",
        path: "/",
        maxAge: 7 * 24 * 60 * 60,
      })
    }
  } catch {}

  return response
}

export const config = {
  matcher: ["/((?!_next/static|_next/image|favicon.ico|.*\\.png$|.*\\.jpg$|.*\\.jpeg$|.*\\.gif$|.*\\.svg$).*)"],
}
