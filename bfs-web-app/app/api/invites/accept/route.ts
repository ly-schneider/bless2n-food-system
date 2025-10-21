import { cookies, headers } from "next/headers"
import { NextResponse } from "next/server"
import { API_BASE_URL } from "@/lib/api"
import { randomUrlSafe } from "@/lib/crypto"
import { resolveCookieNames, setCsrfCookie, setRefreshCookie } from "@/lib/server/cookies"

export async function POST(req: Request) {
  const cookieStore = await cookies()
  const hdrs = await headers()
  const names = await resolveCookieNames()

  const body = (await req.json().catch(() => null)) as { token?: string; firstName?: string; lastName?: string } | null
  if (!body || !body.token || !body.firstName) {
    return NextResponse.json({ error: true, message: "Invalid payload" }, { status: 400 })
  }

  const ua = hdrs.get("user-agent") || ""
  const res = await fetch(`${API_BASE_URL}/v1/invites/accept`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      "X-Internal-Call": "1",
      // Forward UA so backend can label the session
      "X-Forwarded-User-Agent": ua,
    },
    body: JSON.stringify({ token: body.token, firstName: body.firstName, lastName: body.lastName }),
  })

  if (!res.ok) {
    return NextResponse.json({ error: true, message: "Unauthorized" }, { status: res.status })
  }

  const data = (await res.json()) as {
    access_token: string
    expires_in: number
    token_type: string
    user: unknown
    refresh_token?: string
    csrf_token?: string
  }
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
