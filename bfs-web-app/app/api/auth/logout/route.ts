import { cookies, headers } from "next/headers"
import { NextResponse } from "next/server"
import { API_BASE_URL } from "@/lib/api"

export async function POST() {
  const cookieStore = await cookies()
  const hdrs = await headers()
  const proto = (hdrs.get("x-forwarded-proto") || "").toLowerCase()
  const secure = proto === "https"
  const rtName = secure ? "__Host-rt" : "rt"
  const csrfName = secure ? "__Host-csrf" : "csrf"

  const rt = cookieStore.get(rtName)?.value
  const csrfCookie = cookieStore.get(csrfName)?.value
  const csrfHeader = hdrs.get("X-CSRF") || hdrs.get("x-csrf")

  if (!csrfCookie || !csrfHeader || csrfCookie !== csrfHeader) {
    return NextResponse.json({ error: true, message: "Forbidden" }, { status: 403 })
  }

  if (rt) {
    const cookieHeader = [
      `${rtName}=${encodeURIComponent(rt)}`,
      csrfCookie ? `${csrfName}=${encodeURIComponent(csrfCookie)}` : null,
    ]
      .filter(Boolean)
      .join("; ")
    await fetch(`${API_BASE_URL}/v1/auth/logout`, {
      method: "POST",
      headers: { "X-CSRF": csrfHeader!, Cookie: cookieHeader },
    }).catch(() => {})
  }

  // Clear local cookies regardless
  cookieStore.set({ name: rtName, value: "", path: "/", httpOnly: true, secure, sameSite: "lax", maxAge: -1 })
  cookieStore.set({ name: csrfName, value: "", path: "/", httpOnly: false, secure, sameSite: "lax", maxAge: -1 })

  return NextResponse.json({ ok: true })
}
