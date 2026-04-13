"use client"

import { Minus, Plus } from "lucide-react"
import Image from "next/image"
import { useCallback, useState } from "react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import type { ListResponse, ProductDTO } from "@/types"

function StockControl({
  product,
  currentStock,
  token,
  onStockUpdated,
}: {
  product: ProductDTO
  currentStock: number
  token: string
  onStockUpdated: (productId: string, newStock: number) => void
}) {
  const [localStock, setLocalStock] = useState(String(currentStock))
  const [saving, setSaving] = useState(false)

  const adjustStock = useCallback(
    async (delta: number) => {
      setSaving(true)
      try {
        const res = await fetch(`/api/v1/pos/products/${product.id}/inventory`, {
          method: "PATCH",
          headers: {
            Authorization: `Bearer ${token}`,
            "Content-Type": "application/json",
          },
          body: JSON.stringify({ delta, reason: "manual_adjust" }),
        })
        if (res.ok) {
          const data = (await res.json()) as { quantity: number }
          setLocalStock(String(data.quantity))
          onStockUpdated(product.id, data.quantity)
        }
      } catch {
      } finally {
        setSaving(false)
      }
    },
    [product.id, token, onStockUpdated]
  )

  const handleAbsoluteSet = useCallback(async () => {
    const target = parseInt(localStock, 10)
    if (isNaN(target) || target < 0) return
    const delta = target - currentStock
    if (delta === 0) return
    await adjustStock(delta)
  }, [localStock, currentStock, adjustStock])

  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent) => {
      if (e.key === "Enter") handleAbsoluteSet()
    },
    [handleAbsoluteSet]
  )

  return (
    <div className="flex items-center gap-1.5">
      <Button
        variant="outline"
        size="icon"
        className="size-9 shrink-0 rounded-[10px]"
        disabled={saving || currentStock <= 0}
        onClick={() => adjustStock(-1)}
      >
        <Minus className="size-4" />
      </Button>
      <Input
        type="number"
        min={0}
        value={localStock}
        onChange={(e) => setLocalStock(e.target.value)}
        onBlur={handleAbsoluteSet}
        onKeyDown={handleKeyDown}
        disabled={saving}
        className="bg-card h-9 w-16 [appearance:textfield] rounded-[10px] text-center [&::-webkit-inner-spin-button]:appearance-none [&::-webkit-outer-spin-button]:appearance-none"
      />
      <Button
        variant="outline"
        size="icon"
        className="size-9 shrink-0 rounded-[10px]"
        disabled={saving}
        onClick={() => adjustStock(1)}
      >
        <Plus className="size-4" />
      </Button>
    </div>
  )
}

export function InventoryManagement({
  products,
  token,
  onStockUpdated,
}: {
  products: ListResponse<ProductDTO>
  token: string
  onStockUpdated: (productId: string, newStock: number) => void
}) {
  const simpleProducts = products.items.filter((p) => p.type !== "menu" && p.isActive !== false)

  return (
    <div className="h-full space-y-3 overflow-y-auto px-3 pb-4 md:px-4">
      <h1 className="mb-1 text-xl font-semibold">Bestand verwalten</h1>
      <div className="flex flex-col gap-2 md:grid md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
        {simpleProducts.map((product) => {
          const stock = product.availableQuantity ?? 0
          return (
            <div key={product.id} className="flex items-center gap-4 rounded-xl border p-3">
              <div className="relative size-20 shrink-0 overflow-hidden rounded-xl bg-[#cec9c6]">
                {product.image && (
                  <Image
                    src={product.image}
                    alt={product.name}
                    fill
                    sizes="80px"
                    quality={90}
                    className="object-cover"
                    unoptimized={product.image.includes("localhost") || product.image.includes("127.0.0.1")}
                  />
                )}
              </div>
              <div className="min-w-0 flex-1 space-y-1.5">
                <p className="truncate text-sm font-medium">{product.name}</p>
                <StockControl product={product} currentStock={stock} token={token} onStockUpdated={onStockUpdated} />
              </div>
            </div>
          )
        })}
      </div>
      {simpleProducts.length === 0 && (
        <p className="text-muted-foreground py-8 text-center text-sm">Keine Produkte gefunden.</p>
      )}
    </div>
  )
}
