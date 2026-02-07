"use client"

import { useEffect, useRef, useState } from "react"

type InventoryUpdate = {
  productId: string
  newStock: number
  delta: number
  timestamp: string
}

type UseInventoryStreamOptions = {
  onUpdate: (stockMap: Map<string, number>) => void
  enabled?: boolean
}

export function useInventoryStream({ onUpdate, enabled = true }: UseInventoryStreamOptions) {
  const [isConnected, setIsConnected] = useState(false)
  const eventSourceRef = useRef<EventSource | null>(null)
  const onUpdateRef = useRef(onUpdate)

  useEffect(() => {
    onUpdateRef.current = onUpdate
  }, [onUpdate])

  useEffect(() => {
    if (!enabled) {
      return
    }

    const es = new EventSource("/api/v1/inventory/stream")
    eventSourceRef.current = es

    es.addEventListener("inventory-snapshot", (e: MessageEvent) => {
      try {
        const snapshot = JSON.parse(e.data) as Record<string, number>
        onUpdateRef.current(new Map(Object.entries(snapshot)))
      } catch {}
    })

    es.addEventListener("inventory-update", (e: MessageEvent) => {
      try {
        const update = JSON.parse(e.data) as InventoryUpdate
        onUpdateRef.current(new Map([[update.productId, update.newStock]]))
      } catch {}
    })

    es.onopen = () => setIsConnected(true)
    es.onerror = () => setIsConnected(false)

    return () => {
      es.close()
      eventSourceRef.current = null
      setIsConnected(false)
    }
  }, [enabled])

  return { isConnected }
}
