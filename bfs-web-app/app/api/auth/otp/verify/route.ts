import { cookies, headers } from "next/headers"
import { NextResponse } from "next/server"
import { API_BASE_URL } from "@/lib/api"
import { randomUrlSafe } from "@/lib/crypto"
import { resolveCookieNames, setCsrfCookie, setRefreshCookie } from "@/lib/server/cookies"

type VerifyBody = { email: string; otp: string }

export async function POST(req: Request) {
  const hdrs = await headers()
  const names = await resolveCookieNames()
  const body = (await req.json()) as VerifyBody
  const ua = hdrs.get("user-agent") || ""
  const xff = hdrs.get("x-forwarded-for") || ""
  const res = await fetch(`${API_BASE_URL}/v1/auth/otp/verify`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      "X-Internal-Call": "1",
      "X-Forwarded-User-Agent": ua,
      ...(xff ? { "X-Forwarded-For": xff } : {}),
    },
    body: JSON.stringify(body),
  })

  if (!res.ok) {
    return NextResponse.json({ error: true, message: "Invalid code" }, { status: 401 })
  }
  const cookieStore = await cookies()

  const data = (await res.json()) as {
    access_token: string
    expires_in: number
    user: unknown
    token_type: string
    refresh_token?: string
    csrf_token?: string
    roles?: string[]
    is_new?: boolean
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
    roles: data.roles,
    is_new: data.is_new,
  })
}
