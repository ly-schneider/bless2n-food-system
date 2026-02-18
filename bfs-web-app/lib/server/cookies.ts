import { headers } from "next/headers"

export type CookieNames = { secure: boolean; csrfName: string }

export async function resolveCookieNames(): Promise<CookieNames> {
  const hdrs = await headers()
  const proto = (hdrs.get("x-forwarded-proto") || "").toLowerCase()
  const secure = proto === "https"
  return {
    secure,
    csrfName: secure ? "__Host-csrf" : "csrf",
  }
}
