"use client"
import React, { createContext, useCallback, useContext, useEffect, useMemo, useState } from "react"
import type { User } from "@/types"

type AuthState = {
  accessToken: string | null
  user: User | null
  expiresAt: number | null
}

type AuthContextType = AuthState & {
  setAuth: (t: string, expiresIn: number, user?: User) => void
  clearAuth: () => void
  refresh: () => Promise<boolean>
  signOut: () => Promise<void>
  getToken: () => string | null
}

const AuthContext = createContext<AuthContextType | undefined>(undefined)

function getCookie(name: string) {
  if (typeof document === "undefined") return null
  const m = document.cookie.match(new RegExp("(?:^|; )" + name.replace(/([.$?*|{}()\[\]\\/+^])/g, "\\$1") + "=([^;]*)"))
  return m && m[1] ? decodeURIComponent(m[1]!) : null
}

const SESSION_KEY = "bfs.auth.session"
type StoredSession = { token: string; expiresAt: number; user?: User }
function readSession(): StoredSession | null {
  if (typeof window === "undefined") return null
  try {
    const raw = sessionStorage.getItem(SESSION_KEY)
    if (!raw) return null
    const parsed = JSON.parse(raw) as StoredSession
    if (!parsed || !parsed.token || !parsed.expiresAt) return null
    return parsed
  } catch {
    return null
  }
}
function writeSession(s: StoredSession) {
  if (typeof window === "undefined") return
  try {
    sessionStorage.setItem(SESSION_KEY, JSON.stringify(s))
  } catch {}
}
function clearSession() {
  if (typeof window === "undefined") return
  try {
    sessionStorage.removeItem(SESSION_KEY)
  } catch {}
}

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [accessToken, setAccessToken] = useState<string | null>(null)
  const [user, setUser] = useState<User | null>(null)
  const [expiresAt, setExpiresAt] = useState<number | null>(null)
  const [refreshTimer, setRefreshTimer] = useState<ReturnType<typeof setTimeout> | null>(null)

  const setAuth = useCallback((t: string, expiresIn: number, u?: User) => {
    const exp = Date.now() + expiresIn * 1000
    setAccessToken(t)
    setExpiresAt(exp)
    if (u) setUser(u)
    // Persist to sessionStorage for hard reloads
    writeSession({ token: t, expiresAt: exp, user: u })
  }, [])

  const clearAuth = useCallback(() => {
    setAccessToken(null)
    setExpiresAt(null)
    setUser(null)
    if (refreshTimer) clearTimeout(refreshTimer)
    setRefreshTimer(null)
    clearSession()
  }, [])

  const refresh = useCallback(async () => {
    try {
      const res = await fetch("/api/auth/refresh", { method: "POST" })
      if (!res.ok) {
        clearAuth()
        return false
      }
      const data = (await res.json()) as { access_token: string; expires_in: number; user?: User }
      setAuth(data.access_token, data.expires_in, data.user)
      return true
    } catch {
      clearAuth()
      return false
    }
  }, [setAuth, clearAuth])

  const signOut = useCallback(async () => {
    const csrf = getCookie("__Host-csrf") || getCookie("csrf")
    try {
      await fetch("/api/auth/logout", { method: "POST", headers: csrf ? { "X-CSRF": csrf } : {} })
    } catch {}
    clearAuth()
  }, [clearAuth])

  // On mount, try to populate from sessionStorage; fallback to refresh using cookies
  useEffect(() => {
    const s = readSession()
    const now = Date.now()
    if (s && s.token && s.expiresAt && s.expiresAt - now > 5_000) {
      setAccessToken(s.token)
      setExpiresAt(s.expiresAt)
      if (s.user) setUser(s.user)
      return
    }
    void refresh()
  }, [])

  // Schedule token refresh ~60s before expiry
  useEffect(() => {
    if (!expiresAt) return
    const now = Date.now()
    const msUntil = Math.max(0, expiresAt - now - 60_000)
    if (refreshTimer) clearTimeout(refreshTimer)
    const t = setTimeout(() => {
      void refresh()
    }, msUntil)
    setRefreshTimer(t)
    return () => clearTimeout(t)
  }, [expiresAt, refresh])

  const value = useMemo(
    () => ({ accessToken, user, expiresAt, setAuth, clearAuth, refresh, signOut }),
    [accessToken, user, expiresAt, setAuth, clearAuth, refresh, signOut]
  )
  // expose a snapshot getter that always reads latest state
  const getToken = () => accessToken
  const valueWithGetter = useMemo(() => ({ ...value, getToken }), [value])

  return <AuthContext.Provider value={valueWithGetter}>{children}</AuthContext.Provider>
}

export function useAuth() {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error("useAuth must be used within AuthProvider")
  return ctx
}
