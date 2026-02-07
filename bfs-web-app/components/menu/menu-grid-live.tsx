"use client"

import { useCallback, useState } from "react"
import { useInventoryStream } from "@/hooks/use-inventory-stream"
import type { ListResponse, ProductDTO } from "@/types"
import MenuGrid from "./menu-grid"

function applyStockToProduct(p: ProductDTO, stockMap: Map<string, number>): ProductDTO {
  let updated = p

  // Update menu slot options with stock data
  if (p.type === "menu" && p.menu?.slots) {
    updated = {
      ...p,
      menu: {
        ...p.menu,
        slots: p.menu.slots.map((slot) => ({
          ...slot,
          options:
            slot.options?.map((opt) => {
              const optStock = stockMap.get(opt.id)
              if (optStock === undefined) return opt
              return {
                ...opt,
                availableQuantity: optStock,
                isAvailable: optStock > 0,
                isLowStock: optStock > 0 && optStock <= 10,
              }
            }) ?? null,
        })),
      },
    }
  } else {
    // Update simple product stock
    const newStock = stockMap.get(p.id)
    if (newStock !== undefined) {
      updated = {
        ...p,
        availableQuantity: newStock,
        isAvailable: newStock > 0,
        isLowStock: newStock > 0 && newStock <= 10,
      }
    }
  }

  return updated
}

export function MenuGridLive({ initialProducts }: { initialProducts: ListResponse<ProductDTO> }) {
  const [products, setProducts] = useState(initialProducts)

  const handleStockUpdate = useCallback((stockMap: Map<string, number>) => {
    setProducts((prev) => ({
      ...prev,
      items: prev.items.map((p) => applyStockToProduct(p, stockMap)),
    }))
  }, [])

  useInventoryStream({ onUpdate: handleStockUpdate, enabled: true })

  return <MenuGrid products={products} />
}
