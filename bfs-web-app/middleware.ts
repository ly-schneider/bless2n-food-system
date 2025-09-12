import { NextRequest, NextResponse } from "next/server"
import { AuthService } from "./lib/auth"
import { Permission, RBACService } from "./lib/rbac"

export async function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl

  if (
    pathname.startsWith("/_next/") ||
    pathname.startsWith("/api/health") ||
    pathname.includes(".")
  ) {
    return NextResponse.next()
  }

  const { accessToken } = await AuthService.getTokensFromRequest(request)

  // Handle public routes (no auth required)
  const publicRoutes = ["/", "/menu", "/cart", "/checkout", "/login"]
  const isPublicRoute = publicRoutes.some((route) => pathname === route || pathname.startsWith(route))

  if (isPublicRoute) {
    // Create guest session if none exists
    if (!accessToken && !request.cookies.get("guest_id")) {
      const guestSession = await AuthService.createGuestSession()
      const response = NextResponse.next()
      response.cookies.set("guest_id", guestSession.guestId!, {
        httpOnly: true,
        secure: process.env.NODE_ENV === "production",
        sameSite: "lax",
        maxAge: 7 * 24 * 60 * 60,
        path: "/",
      })
      return response
    }
    return NextResponse.next()
  }

  // Handle auth API routes (allow unauthenticated access)
  if (pathname.startsWith("/api/auth/")) {
    return NextResponse.next()
  }

  // Require authentication for protected routes
  if (!accessToken) {
    if (pathname.startsWith("/api/")) {
      return new NextResponse(JSON.stringify({ error: "Authentication required" }), {
        status: 401,
        headers: { "Content-Type": "application/json" },
      })
    }

    // Redirect to login for web routes
    const loginUrl = new URL("/login", request.url)
    loginUrl.searchParams.set("redirect", pathname)
    return NextResponse.redirect(loginUrl)
  }

  // Try to get current user (will handle token refresh if needed)
  const currentUser = await AuthService.getCurrentUser()
  if (!currentUser) {
    if (pathname.startsWith("/api/")) {
      return new NextResponse(JSON.stringify({ error: "Invalid or expired token" }), {
        status: 401,
        headers: { "Content-Type": "application/json" },
      })
    }

    // Clear invalid tokens and redirect to login
    const response = NextResponse.redirect(new URL("/login", request.url))
    response.cookies.delete("access_token")
    response.cookies.delete("refresh_token")
    return response
  }

  // Handle admin routes
  if (pathname.startsWith("/admin") || pathname.startsWith("/api/admin/")) {
    if (!RBACService.canAccessRoute(currentUser.role, pathname)) {
      const error = pathname.startsWith("/api/")
        ? new NextResponse(JSON.stringify({ error: "Admin access required" }), {
            status: 403,
            headers: { "Content-Type": "application/json" },
          })
        : NextResponse.redirect(new URL("/?error=unauthorized", request.url))

      return error
    }
  }

  // Handle POS routes
  if (pathname.startsWith("/pos") || pathname.startsWith("/api/pos/")) {
    if (!RBACService.hasPermission(currentUser.role, Permission.POS_ACCESS)) {
      const error = pathname.startsWith("/api/")
        ? new NextResponse(JSON.stringify({ error: "POS access required" }), {
            status: 403,
            headers: { "Content-Type": "application/json" },
          })
        : NextResponse.redirect(new URL("/?error=unauthorized", request.url))

      return error
    }
  }

  // Set security headers for all responses
  const response = NextResponse.next()

  // Security headers
  response.headers.set("X-Frame-Options", "DENY")
  response.headers.set("X-Content-Type-Options", "nosniff")
  response.headers.set("Referrer-Policy", "origin-when-cross-origin")

  // CSP optimized for WebView
  const csp = [
    "default-src 'self'",
    "script-src 'self' 'unsafe-inline' 'unsafe-eval'", // WebView may require unsafe-inline
    "style-src 'self' 'unsafe-inline'",
    "img-src 'self' data: blob: https:",
    "font-src 'self' data:",
    "connect-src 'self' ws: wss:",
    "frame-src 'none'",
    "object-src 'none'",
    "base-uri 'self'",
  ].join("; ")

  response.headers.set("Content-Security-Policy", csp)

  return response
}

export const config = {
  matcher: [
    "/((?!_next/static|_next/image|favicon.ico|.*\\.png$|.*\\.jpg$|.*\\.jpeg$|.*\\.gif$|.*\\.svg$).*)",
  ],
}
