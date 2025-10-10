import { cookies, headers } from "next/headers"
import { NextResponse } from "next/server"
import { API_BASE_URL } from "@/lib/api"

export async function GET(req: Request) {
  const cookieStore = await cookies()
  const hdrs = await headers()
  const proto = (hdrs.get("x-forwarded-proto") || "").toLowerCase()
  const secure = proto === "https"
  const rtName = secure ? "__Host-rt" : "rt"
  const csrfName = secure ? "__Host-csrf" : "csrf"

  const url = new URL(req.url)
  const origin = hdrs.get("x-forwarded-host")
    ? `${secure ? "https" : "http"}://${hdrs.get("x-forwarded-host")}`
    : url.origin
  const mkAbs = (p: string) => new URL(p, origin)
  const code = url.searchParams.get("code")
  const state = url.searchParams.get("state")
  const stateCookie = cookieStore.get("g_oauth_state")?.value
  const verifier = cookieStore.get("g_oauth_verifier")?.value
  const nextRaw = cookieStore.get("g_oauth_next")?.value
  const next = nextRaw ? decodeURIComponent(nextRaw) : "/"
  const redirectUri = `${
    hdrs.get("x-forwarded-host")
      ? (secure ? "https" : "http") + "://" + hdrs.get("x-forwarded-host")
      : new URL(req.url).origin
  }/api/auth/google/callback`

  if (!code || !state || !stateCookie || state !== stateCookie || !verifier) {
    return NextResponse.redirect(mkAbs(`/login?error=oauth_google`))
  }

  const res = await fetch(`${API_BASE_URL}/v1/auth/google/code`, {
    method: "POST",
    headers: { "Content-Type": "application/json", "X-Internal-Call": "1" },
    body: JSON.stringify({
      code,
      code_verifier: verifier,
      redirect_uri: redirectUri,
      nonce: cookieStore.get("g_oauth_nonce")?.value || "",
    }),
  })
  if (!res.ok) {
    return NextResponse.redirect(mkAbs(`/login?error=oauth_google`))
  }
  const data = (await res.json()) as {
    access_token: string
    expires_in: number
    token_type: string
    user: unknown
    refresh_token?: string
    csrf_token?: string
  }
  // Set cookies
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
  const csrf = data.csrf_token || randomURLSafe(16)
  cookieStore.set({
    name: csrfName,
    value: csrf,
    httpOnly: false,
    secure,
    sameSite: "lax",
    path: "/",
    maxAge: 7 * 24 * 60 * 60,
  })

  // clear temp cookies
  cookieStore.set({ name: "g_oauth_state", value: "", httpOnly: true, secure, sameSite: "lax", path: "/", maxAge: -1 })
  cookieStore.set({ name: "g_oauth_nonce", value: "", httpOnly: true, secure, sameSite: "lax", path: "/", maxAge: -1 })
  cookieStore.set({
    name: "g_oauth_verifier",
    value: "",
    httpOnly: true,
    secure,
    sameSite: "lax",
    path: "/",
    maxAge: -1,
  })
  cookieStore.set({ name: "g_oauth_next", value: "", httpOnly: true, secure, sameSite: "lax", path: "/", maxAge: -1 })

  const safeNext = next && !/^https?:/i.test(next) ? next : "/"
  return NextResponse.redirect(mkAbs(safeNext))
}

function randomURLSafe(n: number) {
  const bytes = crypto.getRandomValues(new Uint8Array(n))
  const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_" // URL-safe
  let out = ""
  for (let i = 0; i < bytes.length; i++) {
    const b = bytes[i]!
    out += chars[b % chars.length]
  }
  return out
}
