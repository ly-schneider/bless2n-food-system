"use client"

import { useCallback } from "react"
import { useRouter } from "next/navigation"
import { getDeviceToken, clearDeviceToken, type DeviceType } from "@/lib/device-auth"

/**
 * Device fetch hook for API requests with device JWT authentication.
 *
 * Authentication flow:
 * 1. Device requests access (POST /v1/pos/requests or /v1/stations/requests)
 * 2. Admin approves the device and receives a JWT token
 * 3. Device stores the token via setDeviceToken()
 * 4. This hook attaches the Bearer token to all API requests
 * 5. On 401/403, clears token and redirects to setup page
 *
 * Usage:
 * ```tsx
 * const deviceFetch = useDeviceFetch("pos")
 *
 * const response = await deviceFetch("/v1/pos/orders", {
 *   method: "POST",
 *   body: JSON.stringify({ items: [...] }),
 * })
 * ```
 */
export function useDeviceFetch(deviceType: DeviceType) {
  const router = useRouter()
  const token = getDeviceToken()

  const setupPath = deviceType === "pos" ? "/pos/setup" : "/station/setup"

  return useCallback(
    async (input: RequestInfo | URL, init?: RequestInit): Promise<Response> => {
      const url = typeof input === "string" ? input : (input as URL).toString()
      const headers = new Headers(init?.headers || {})

      if (token) {
        headers.set("Authorization", `Bearer ${token}`)
      }

      const res = await fetch(url, { ...init, headers })

      // Handle auth errors by clearing token and redirecting
      if (res.status === 401 || res.status === 403) {
        clearDeviceToken()
        router.push(setupPath)
      }

      return res
    },
    [token, router, setupPath]
  )
}

/**
 * Device fetch without React hooks - for use outside components.
 *
 * Usage:
 * ```ts
 * const response = await deviceFetch("/v1/pos/orders", {
 *   method: "POST",
 *   body: JSON.stringify({ items: [...] }),
 * })
 * ```
 */
export async function deviceFetch(input: RequestInfo | URL, init?: RequestInit): Promise<Response> {
  const token = getDeviceToken()
  const url = typeof input === "string" ? input : (input as URL).toString()
  const headers = new Headers(init?.headers || {})

  if (token) {
    headers.set("Authorization", `Bearer ${token}`)
  }

  return fetch(url, { ...init, headers })
}
