"use client"

import { useEffect } from "react"

function isInstalledDisplayMode(): boolean {
  if (typeof window === "undefined") return false
  if (window.matchMedia("(display-mode: fullscreen)").matches) return true
  if (window.matchMedia("(display-mode: standalone)").matches) return true
  const nav = window.navigator as Navigator & { standalone?: boolean }
  return nav.standalone === true
}

export function PWARuntime() {
  useEffect(() => {
    if (!isInstalledDisplayMode()) return
    const orientation = screen.orientation as ScreenOrientation & { lock?: (o: string) => Promise<void> }
    orientation.lock?.("portrait").catch(() => {})
  }, [])

  return null
}
