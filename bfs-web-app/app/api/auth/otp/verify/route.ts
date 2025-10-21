import { cookies, headers } from "next/headers"
import { NextResponse } from "next/server"
import { API_BASE_URL } from "@/lib/api"

type VerifyBody = { email: string; otp: string }

export async function POST(req: Request) {
  const hdrs = await headers()
  const proto = (hdrs.get("x-forwarded-proto") || "").toLowerCase()
  const secure = proto === "https"
  const rtName = secure ? "__Host-rt" : "rt"
  const csrfName = secure ? "__Host-csrf" : "csrf"
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

  console.log("OTP verify request body:", body)
  console.log("OTP verify response status:", res.status)

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
    roles: data.roles,
    is_new: data.is_new,
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
