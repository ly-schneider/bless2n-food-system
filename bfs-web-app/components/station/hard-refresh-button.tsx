"use client"

import { RefreshCwIcon } from "lucide-react"
import { useCallback, useState } from "react"
import { Button } from "@/components/ui/button"

export function HardRefreshButton() {
  const [refreshing, setRefreshing] = useState(false)

  const hardRefresh = useCallback(async () => {
    setRefreshing(true)
    try {
      if ("serviceWorker" in navigator) {
        const reg = await navigator.serviceWorker.getRegistration("/station")
        if (reg?.active) {
          reg.active.postMessage("hard-refresh")
          await new Promise((r) => setTimeout(r, 200))
        }
        if (reg) await reg.unregister()
      }
      if ("caches" in window) {
        const keys = await caches.keys()
        await Promise.all(keys.map((k) => caches.delete(k)))
      }
    } catch {}
    window.location.reload()
  }, [])

  return (
    <Button
      type="button"
      variant="outline"
      size="sm"
      onClick={hardRefresh}
      disabled={refreshing}
      aria-label="Aktualisieren"
      className="gap-1.5 px-2 sm:px-3"
    >
      <RefreshCwIcon className={`size-4 ${refreshing ? "animate-spin" : ""}`} />
      <span className="hidden sm:inline">Aktualisieren</span>
    </Button>
  )
}
