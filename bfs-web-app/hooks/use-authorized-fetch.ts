"use client"
import { useAuth } from '@/contexts/auth-context'
import { API_BASE_URL } from '@/lib/api'
import { useCallback } from 'react'

/**
 * Authorized fetch hook for API requests with automatic token refresh and CSRF protection.
 * 
 * Authentication flow:
 * 1. Read-only requests (GET/HEAD/OPTIONS): Direct API calls with Bearer token
 * 2. Mutating requests (POST/PATCH/DELETE): Routed through proxy for CSRF validation
 * 3. On 401 responses: Automatically refresh tokens and retry request
 * 
 * CSRF protection:
 * - Mutating requests require CSRF token (double-submit cookie pattern)
 * - Token is read from browser cookie and sent as X-CSRF header
 * - Proxy validates header matches cookie before forwarding to backend
 */

function getCSRFCookie(): string | null {
  if (typeof document === 'undefined') return null
  const name = (document.location.protocol === 'https:' ? '__Host-' : '') + 'csrf'
  const m = document.cookie.match(new RegExp('(?:^|; )' + name.replace(/([.$?*|{}()\[\]\\/+^])/g, '\\$1') + '=([^;]*)'))
  return m && m[1] ? decodeURIComponent(m[1]!) : null
}

export function useAuthorizedFetch() {
  const { accessToken, refresh, getToken } = useAuth()

  return useCallback(async (input: RequestInfo | URL, init?: RequestInit) => {
    const originalUrl = typeof input === 'string' ? input : (input as URL).toString()
    const method = (init?.method || 'GET').toUpperCase()
    const headers = new Headers(init?.headers || {})
    if (accessToken) headers.set('Authorization', `Bearer ${accessToken}`)

    const isMutating = !(method === 'GET' || method === 'HEAD' || method === 'OPTIONS')
    const isApiCall = typeof originalUrl === 'string' && originalUrl.startsWith(String(API_BASE_URL))

    async function doFetch() {
      // For mutating API requests, use proxy to handle CSRF validation
      if (isMutating && isApiCall) {
        const csrf = getCSRFCookie()
        if (!csrf) {
          // No CSRF token available, trigger refresh by returning 401
          return new Response('{"error":true,"message":"No CSRF token"}', { 
            status: 401, 
            headers: { 'Content-Type': 'application/json' } 
          })
        }

        // Prepare request payload for proxy
        const forwardBody = typeof init?.body === 'string' || init?.body == null ? init?.body : undefined
        const forwardHeaders: Record<string, string> = {}
        headers.forEach((v, k) => { forwardHeaders[k] = v })

        return fetch('/api/proxy', {
          method: 'POST',
          headers: { 
            'Content-Type': 'application/json', 
            'X-CSRF': csrf,
            ...(accessToken ? { 'Authorization': `Bearer ${accessToken}` } : {})
          },
          body: JSON.stringify({ 
            url: originalUrl, 
            method, 
            headers: forwardHeaders, 
            body: forwardBody ?? null 
          }),
        })
      }

      // For read-only requests, call API directly
      return fetch(originalUrl, { ...init, headers })
    }

    let res = await doFetch()
    if (res.status === 401) {
      const ok = await refresh()
      if (ok) {
        const token = getToken()
        if (token) headers.set('Authorization', `Bearer ${token}`)
        res = await doFetch()
      }
    }
    return res
  }, [accessToken, refresh, getToken])
}
