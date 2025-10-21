import { cookies, headers } from "next/headers"
import { NextResponse } from "next/server"
import { randomUrlSafe, sha256Base64Url } from "@/lib/crypto"
import { resolveCookieNames } from "@/lib/server/cookies"

export async function GET(req: Request) {
  const url = new URL(req.url)
  const next = url.searchParams.get("next") || "/"
  const hdrs = await headers()
  const { secure } = await resolveCookieNames()
  const cookieStore = await cookies()
  const origin = hdrs.get("x-forwarded-host")
    ? `${secure ? "https" : "http"}://${hdrs.get("x-forwarded-host")}`
    : new URL(req.url).origin

  const clientId = process.env.NEXT_PUBLIC_GOOGLE_CLIENT_ID
  if (!clientId) return NextResponse.json({ error: true, message: "Missing Google client id" }, { status: 500 })
  const redirectUri = `${origin}/api/auth/google/callback`

  const state = randomUrlSafe(24)
  const nonce = randomUrlSafe(24)
  const verifier = randomUrlSafe(64)
  const challenge = await sha256Base64Url(verifier)

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
