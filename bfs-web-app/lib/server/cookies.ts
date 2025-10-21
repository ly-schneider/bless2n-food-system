import { headers } from "next/headers"

type SameSite = "lax" | "strict" | "none"

export type CookieRecord = { value?: string }

export interface CookieStore {
  set: (args: {
    name: string
    value: string
    path: string
    httpOnly: boolean
    secure: boolean
    sameSite: SameSite
    maxAge: number
  }) => void
  get: (name: string) => CookieRecord | undefined
}

export type CookieNames = { secure: boolean; rtName: string; csrfName: string }

export async function resolveCookieNames(): Promise<CookieNames> {
  const hdrs = await headers()
  const proto = (hdrs.get("x-forwarded-proto") || "").toLowerCase()
  const secure = proto === "https"
  return {
    secure,
    rtName: secure ? "__Host-rt" : "rt",
    csrfName: secure ? "__Host-csrf" : "csrf",
  }
}

export function clearAuthCookies(cookieStore: CookieStore, names: CookieNames) {
  cookieStore.set({
    name: names.rtName,
    value: "",
    path: "/",
    httpOnly: true,
    secure: names.secure,
    sameSite: "lax",
    maxAge: -1,
  })
  cookieStore.set({
    name: names.csrfName,
    value: "",
    path: "/",
    httpOnly: false,
    secure: names.secure,
    sameSite: "lax",
    maxAge: -1,
  })
}

export function setRefreshCookie(cookieStore: CookieStore, names: CookieNames, refreshToken: string) {
  cookieStore.set({
    name: names.rtName,
    value: refreshToken,
    httpOnly: true,
    secure: names.secure,
    sameSite: "lax",
    path: "/",
    maxAge: 7 * 24 * 60 * 60,
  })
}

export function setCsrfCookie(cookieStore: CookieStore, names: CookieNames, csrfToken: string) {
  cookieStore.set({
    name: names.csrfName,
    value: csrfToken,
    httpOnly: false,
    secure: names.secure,
    sameSite: "lax",
    path: "/",
    maxAge: 7 * 24 * 60 * 60,
  })
}
