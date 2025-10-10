import { cookies, headers } from "next/headers"
import { NextResponse } from "next/server"
import { API_BASE_URL } from "@/lib/api"

export async function POST(req: Request) {
  const cookieStore = await cookies()
  const hdrs = await headers()
  const proto = (hdrs.get("x-forwarded-proto") || "").toLowerCase()
  const secure = proto === "https"
  const rtName = secure ? "__Host-rt" : "rt"
  const csrfName = secure ? "__Host-csrf" : "csrf"

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
  if (data.refresh_token) {
    cookieStore.set({
      name: rtName,
      value: data.refresh_token,
      httpOnly: true,
      secure,
      sameSite: "lax",
      path: "/",
      maxAge: 7 * 24 * 60 * 60,
    })
  }
  const csrf = data.csrf_token || generateRandom(16)
  cookieStore.set({
    name: csrfName,
    value: csrf,
    httpOnly: false,
    secure,
    sameSite: "lax",
    path: "/",
    maxAge: 7 * 24 * 60 * 60,
  })

  return NextResponse.json({
    access_token: data.access_token,
    expires_in: data.expires_in,
    token_type: data.token_type,
    user: data.user,
  })
}

function generateRandom(n: number) {
  const bytes = crypto.getRandomValues(new Uint8Array(n))
  const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_" // URL-safe
  let out = ""
  for (let i = 0; i < bytes.length; i++) {
    const b = bytes[i] ?? 0
    out += chars[b % chars.length]
  }
  return out
}
