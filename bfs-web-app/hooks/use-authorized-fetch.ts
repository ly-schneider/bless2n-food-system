"use client"
import { useCallback } from "react"
import { useAuth } from "@/contexts/auth-context"
import { API_BASE_URL } from "@/lib/api"
import { getCSRFToken } from "@/lib/csrf"

/**
 * Authorized fetch hook for API requests with automatic token refresh and CSRF protection.
 *
 * Authentication flow:
 * 1. Read-only requests (GET/HEAD/OPTIONS): Direct API calls with Bearer token
 * 2. Mutating requests (POST/PATCH/DELETE): Adds X-CSRF header automatically
 * 3. On 401 responses: Automatically refresh tokens and retry request
 *
 * CSRF protection:
 * - Mutating requests require CSRF token (double-submit cookie pattern)
 * - Token is read from browser cookie and sent as X-CSRF header
 * - Next.js catch-all API route validates header matches cookie before forwarding
 */

export function useAuthorizedFetch() {
  const { accessToken, refresh, getToken } = useAuth()

  return useCallback(
    async (input: RequestInfo | URL, init?: RequestInit) => {
      const originalUrl = typeof input === "string" ? input : (input as URL).toString()
      const method = (init?.method || "GET").toUpperCase()
      const headers = new Headers(init?.headers || {})
      if (accessToken) headers.set("Authorization", `Bearer ${accessToken}`)

      const isMutating = !(method === "GET" || method === "HEAD" || method === "OPTIONS")
      const isApiCall = typeof originalUrl === "string" && originalUrl.startsWith(String(API_BASE_URL))

      async function doFetch() {
        // For mutating API requests to our API, ensure CSRF header is set
        if (isMutating && isApiCall && !headers.has("X-CSRF") && !headers.has("x-csrf")) {
          let csrf = getCSRFToken()
          if (!csrf) {
            // Try to refresh to obtain a fresh CSRF cookie
            const ok = await refresh()
            if (ok) csrf = getCSRFToken()
          }
          if (csrf) headers.set("X-CSRF", csrf)
        }
        return fetch(originalUrl, { ...init, headers })
      }

      let res = await doFetch()
      if (res.status === 401) {
        const ok = await refresh()
        if (ok) {
          const token = getToken()
          if (token) headers.set("Authorization", `Bearer ${token}`)
          res = await doFetch()
        }
      }
      return res
    },
    [accessToken, refresh, getToken]
  )
}
