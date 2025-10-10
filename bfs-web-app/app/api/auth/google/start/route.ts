import { cookies, headers } from "next/headers"
import { NextResponse } from "next/server"

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

export async function GET(req: Request) {
  const url = new URL(req.url)
  const next = url.searchParams.get("next") || "/"
  const hdrs = await headers()
  const proto = (hdrs.get("x-forwarded-proto") || "").toLowerCase()
  const secure = proto === "https"
  const cookieStore = await cookies()
  const origin = hdrs.get("x-forwarded-host")
    ? `${secure ? "https" : "http"}://${hdrs.get("x-forwarded-host")}`
    : new URL(req.url).origin

  const clientId = process.env.NEXT_PUBLIC_GOOGLE_CLIENT_ID
  if (!clientId) return NextResponse.json({ error: true, message: "Missing Google client id" }, { status: 500 })
  const redirectUri = `${origin}/api/auth/google/callback`

  const state = randomURLSafe(24)
  const nonce = randomURLSafe(24)
  const verifier = randomURLSafe(64)
  // S256
  const challengeBuf = await crypto.subtle.digest("SHA-256", new TextEncoder().encode(verifier))
  // Avoid spread on typed arrays to support stricter TS configs
  const challengeBase = Array.from(new Uint8Array(challengeBuf))
    .map((b) => String.fromCharCode(b))
    .join("")
  const challenge = btoa(challengeBase).replaceAll("+", "-").replaceAll("/", "_").replace(/=+$/, "")

  // temp cookies (5 minutes)
  const opts = { httpOnly: true, secure, sameSite: "lax" as const, path: "/", maxAge: 300 }
  cookieStore.set({ name: "g_oauth_state", value: state, ...opts })
  cookieStore.set({ name: "g_oauth_nonce", value: nonce, ...opts })
  cookieStore.set({ name: "g_oauth_verifier", value: verifier, ...opts })
  cookieStore.set({ name: "g_oauth_next", value: encodeURIComponent(next), ...opts })

  const authURL = new URL("https://accounts.google.com/o/oauth2/v2/auth")
  authURL.searchParams.set("response_type", "code")
  authURL.searchParams.set("client_id", clientId)
  authURL.searchParams.set("redirect_uri", redirectUri)
  authURL.searchParams.set("scope", "openid email profile")
  authURL.searchParams.set("state", state)
  authURL.searchParams.set("nonce", nonce)
  authURL.searchParams.set("code_challenge", challenge)
  authURL.searchParams.set("code_challenge_method", "S256")
  authURL.searchParams.set("prompt", "select_account")

  return NextResponse.redirect(authURL.toString())
}
