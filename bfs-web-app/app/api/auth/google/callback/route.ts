import { cookies, headers } from "next/headers"
import { NextResponse } from "next/server"
import { API_BASE_URL } from "@/lib/api"
import { randomUrlSafe } from "@/lib/crypto"
import { resolveCookieNames, setCsrfCookie, setRefreshCookie } from "@/lib/server/cookies"

export async function GET(req: Request) {
  const cookieStore = await cookies()
  const hdrs = await headers()
  const names = await resolveCookieNames()
  const secure = names.secure

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
  if (data.refresh_token) setRefreshCookie(cookieStore, names, data.refresh_token)
  const csrf = data.csrf_token || randomUrlSafe(16)
  setCsrfCookie(cookieStore, names, csrf)

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
