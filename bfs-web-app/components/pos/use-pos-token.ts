"use client"

import { useMemo } from "react"

export function usePosToken(): string {
  function randKey(): string {
    const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_"
    const bytes = new Uint8Array(24)
    if (typeof crypto !== "undefined" && crypto.getRandomValues) crypto.getRandomValues(bytes)
    return Array.from(bytes, (b) => chars[b % chars.length]).join("")
  }

  return useMemo(() => {
    if (typeof window === "undefined") return ""
    let k = localStorage.getItem("bfs.posToken")
    if (!k) {
      k = `pos_${randKey()}`
      localStorage.setItem("bfs.posToken", k)
    }
    return k
  }, [])
}
