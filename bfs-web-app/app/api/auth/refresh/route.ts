import { cookies } from "next/headers"
import { NextResponse } from "next/server"
import { API_BASE_URL } from "@/lib/api"
import { randomUrlSafe } from "@/lib/crypto"
import { clearAuthCookies, resolveCookieNames, setCsrfCookie, setRefreshCookie } from "@/lib/server/cookies"

export async function POST(_req: Request) {
  const cookieStore = await cookies()
  const names = await resolveCookieNames()
  // Read refresh token from either secure or non-secure cookie name
  const rt = cookieStore.get(names.rtName)?.value || cookieStore.get(names.rtName === "__Host-rt" ? "rt" : "__Host-rt")?.value

  if (!rt) {
    return NextResponse.json({ error: true, message: "Unauthorized" }, { status: 401 })
  }

  // Call backend refresh - no CSRF required since backend generates new CSRF tokens
  const res = await fetch(`${API_BASE_URL}/v1/auth/refresh`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      "X-Internal-Call": "1",
      // Forward refresh cookie for validation
      Cookie: `${names.rtName}=${encodeURIComponent(rt)}`,
    },
  })

  if (!res.ok) {
    clearAuthCookies(cookieStore, names)
    return NextResponse.json({ error: true, message: "Unauthorized" }, { status: 401 })
  }

  const data = (await res.json()) as {
    access_token: string
    expires_in: number
    token_type: string
    user: unknown
    refresh_token?: string
    csrf_token?: string
  }

  // Set cookies directly from response body (reliable for internal calls)
  if (data.refresh_token) setRefreshCookie(cookieStore, names, data.refresh_token)
  const csrf = data.csrf_token || randomUrlSafe(16)
  setCsrfCookie(cookieStore, names, csrf)

  return NextResponse.json({
    access_token: data.access_token,
    expires_in: data.expires_in,
    token_type: data.token_type,
    user: data.user,
  })
}
