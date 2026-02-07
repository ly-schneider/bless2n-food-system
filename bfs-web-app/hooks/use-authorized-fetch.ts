"use client"

import { useCallback } from "react"
import { useAuth } from "@/contexts/auth-context"

export function useAuthorizedFetch() {
  const { accessToken, getToken } = useAuth()

  return useCallback(
    async (input: RequestInfo | URL, init?: RequestInit) => {
      const originalUrl = typeof input === "string" ? input : (input as URL).toString()
      const headers = new Headers(init?.headers || {})

      // Get the latest token (Better Auth manages refresh automatically)
      const token = getToken() || accessToken
      if (token) {
        headers.set("Authorization", `Bearer ${token}`)
      }

      // Include credentials to send cookies (Better Auth session cookie)
      return fetch(originalUrl, {
        ...init,
        headers,
        credentials: "include",
      })
    },
    [accessToken, getToken]
  )
}
