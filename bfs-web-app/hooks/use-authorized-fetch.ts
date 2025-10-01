"use client"
import { useAuth } from '@/contexts/auth-context'

export function useAuthorizedFetch() {
  const { accessToken, refresh, getToken } = useAuth()
  return async (input: RequestInfo | URL, init?: RequestInit) => {
    const headers = new Headers(init?.headers || {})
    if (accessToken) headers.set('Authorization', `Bearer ${accessToken}`)
    let res = await fetch(input, { ...init, headers })
    if (res.status === 401) {
      const ok = await refresh()
      if (ok) {
        const h2 = new Headers(init?.headers || {})
        const token = getToken()
        if (token) h2.set('Authorization', `Bearer ${token}`)
        res = await fetch(input, { ...init, headers: h2 })
      }
    }
    return res
  }
}

