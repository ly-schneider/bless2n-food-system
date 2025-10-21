import { cookies, headers } from "next/headers"
import { NextResponse } from "next/server"
import { API_BASE_URL } from "@/lib/api"
import { clearAuthCookies, resolveCookieNames } from "@/lib/server/cookies"

export async function POST() {
  const cookieStore = await cookies()
  const hdrs = await headers()
  const names = await resolveCookieNames()

  const rt = cookieStore.get(names.rtName)?.value
  const csrfCookie = cookieStore.get(names.csrfName)?.value
  const csrfHeader = hdrs.get("X-CSRF") || hdrs.get("x-csrf")

  if (!csrfCookie || !csrfHeader || csrfCookie !== csrfHeader) {
    return NextResponse.json({ error: true, message: "Forbidden" }, { status: 403 })
  }

  if (rt) {
    const cookieHeader = [
      `${names.rtName}=${encodeURIComponent(rt)}`,
      csrfCookie ? `${names.csrfName}=${encodeURIComponent(csrfCookie)}` : null,
    ]
      .filter(Boolean)
      .join("; ")
    await fetch(`${API_BASE_URL}/v1/auth/logout`, {
      method: "POST",
      headers: { "X-CSRF": csrfHeader!, Cookie: cookieHeader },
    }).catch(() => {})
  }

  // Clear local cookies regardless
  clearAuthCookies(cookieStore, names)

  return NextResponse.json({ ok: true })
}
