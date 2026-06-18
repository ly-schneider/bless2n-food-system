"use client"

import { useCallback, useState } from "react"
import { useInventoryStream } from "@/hooks/use-inventory-stream"
import type { CartItem } from "@/types/cart"

export function useStockMap(enabled = true): Map<string, number> {
  const [stockMap, setStockMap] = useState<Map<string, number>>(new Map())

  const handleUpdate = useCallback((update: Map<string, number>) => {
    setStockMap((prev) => {
      const next = new Map(prev)
      update.forEach((value, key) => next.set(key, value))
      return next
    })
  }, [])

  useInventoryStream({ onUpdate: handleUpdate, enabled })

  return stockMap
}

export function getStockCap(item: CartItem, stockMap?: Map<string, number>): number | null {
  if (!stockMap || item.product.type === "menu") return null
  return stockMap.get(item.product.id) ?? null
}
