import { NextRequest, NextResponse } from "next/server"

export async function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl

  if (pathname.startsWith("/_next/") || pathname.startsWith("/api/health") || pathname.includes(".")) {
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

  return response
}

export const config = {
  matcher: ["/((?!_next/static|_next/image|favicon.ico|.*\\.png$|.*\\.jpg$|.*\\.jpeg$|.*\\.gif$|.*\\.svg$).*)"],
}
