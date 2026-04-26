"use client"

import { Minus, Plus } from "lucide-react"
import Image from "next/image"
import { memo, useCallback, useRef, useState } from "react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Switch } from "@/components/ui/switch"
import { setProductActive } from "@/lib/api/products"
import { cn } from "@/lib/utils"
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
  const pendingRef = useRef<AbortController | null>(null)

  const adjustStock = useCallback(
    async (delta: number) => {
      pendingRef.current?.abort()

      const optimistic = Math.max(0, parseInt(localStock, 10) + delta)
      setLocalStock(String(optimistic))
      onStockUpdated(product.id, optimistic)

      const controller = new AbortController()
      pendingRef.current = controller
      setSaving(true)
      try {
        const res = await fetch(`/api/v1/pos/products/${product.id}/inventory`, {
          method: "PATCH",
          headers: {
            Authorization: `Bearer ${token}`,
            "Content-Type": "application/json",
          },
          body: JSON.stringify({ delta, reason: "manual_adjust" }),
          signal: controller.signal,
        })
        if (res.ok) {
          const data = (await res.json()) as { quantity: number }
          setLocalStock(String(data.quantity))
          onStockUpdated(product.id, data.quantity)
        } else {
          setLocalStock(String(currentStock))
          onStockUpdated(product.id, currentStock)
        }
      } catch (e) {
        if (e instanceof DOMException && e.name === "AbortError") return
        setLocalStock(String(currentStock))
        onStockUpdated(product.id, currentStock)
      } finally {
        setSaving(false)
      }
    },
    [product.id, token, onStockUpdated, localStock, currentStock]
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
        disabled={saving || parseInt(localStock, 10) <= 0}
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

function ActiveToggle({
  product,
  token,
  onActiveUpdated,
}: {
  product: ProductDTO
  token: string
  onActiveUpdated: (productId: string, isActive: boolean) => void
}) {
  const isActive = product.isActive !== false
  const [saving, setSaving] = useState(false)

  const handleChange = useCallback(
    async (checked: boolean) => {
      setSaving(true)
      onActiveUpdated(product.id, checked)
      try {
        await setProductActive(token, product.id, checked)
      } catch {
        onActiveUpdated(product.id, !checked)
      } finally {
        setSaving(false)
      }
    },
    [product.id, token, onActiveUpdated]
  )

  return (
    <div className="flex items-center gap-2">
      <span className="text-muted-foreground text-xs font-medium">Aktiv</span>
      <Switch checked={isActive} disabled={saving} onCheckedChange={handleChange} />
    </div>
  )
}

const ItemCard = memo(function ItemCard({
  product,
  token,
  onStockUpdated,
  onActiveUpdated,
}: {
  product: ProductDTO
  token: string
  onStockUpdated: (productId: string, newStock: number) => void
  onActiveUpdated: (productId: string, isActive: boolean) => void
}) {
  const isActive = product.isActive !== false
  const isMenu = product.type === "menu"
  const stock = product.availableQuantity ?? 0

  return (
    <div className={cn("flex items-center gap-4 rounded-xl border p-3", !isActive && "opacity-60")}>
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
        {!isActive && (
          <div className="absolute inset-0 flex items-center justify-center bg-black/30">
            <span className="rounded-full bg-black/70 px-2 py-0.5 text-[10px] font-semibold text-white">Inaktiv</span>
          </div>
        )}
      </div>
      <div className="flex min-w-0 flex-1 flex-col gap-2">
        <p className="truncate text-sm font-medium">{product.name}</p>
        <div className="flex flex-wrap items-center justify-between gap-2">
          {isMenu ? (
            <span className="text-muted-foreground text-xs">Menü</span>
          ) : (
            <StockControl product={product} currentStock={stock} token={token} onStockUpdated={onStockUpdated} />
          )}
          <ActiveToggle product={product} token={token} onActiveUpdated={onActiveUpdated} />
        </div>
      </div>
    </div>
  )
})

export function InventoryManagement({
  products,
  token,
  onStockUpdated,
  onActiveUpdated,
}: {
  products: ListResponse<ProductDTO>
  token: string
  onStockUpdated: (productId: string, newStock: number) => void
  onActiveUpdated: (productId: string, isActive: boolean) => void
}) {
  const simpleProducts = products.items.filter((p) => p.type !== "menu")
  const menus = products.items.filter((p) => p.type === "menu")

  const renderGrid = (items: ProductDTO[]) => (
    <div className="flex flex-col gap-2 md:grid md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
      {items.map((product) => (
        <ItemCard
          key={product.id}
          product={product}
          token={token}
          onStockUpdated={onStockUpdated}
          onActiveUpdated={onActiveUpdated}
        />
      ))}
    </div>
  )

  return (
    <div className="h-full space-y-6 overflow-y-auto px-3 pb-4 md:px-4">
      <div className="space-y-3">
        <h1 className="text-xl font-semibold">Produkte verwalten</h1>
        {simpleProducts.length > 0 ? (
          renderGrid(simpleProducts)
        ) : (
          <p className="text-muted-foreground py-4 text-center text-sm">Keine Produkte gefunden.</p>
        )}
      </div>

      <div className="space-y-3">
        <h2 className="text-xl font-semibold">Menüs verwalten</h2>
        {menus.length > 0 ? (
          renderGrid(menus)
        ) : (
          <p className="text-muted-foreground py-4 text-center text-sm">Keine Menüs gefunden.</p>
        )}
      </div>
    </div>
  )
}
