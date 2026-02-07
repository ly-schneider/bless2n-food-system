"use client"

import { useMemo, useState } from "react"
import { ProductCardPOS } from "@/components/pos/product-card-pos"
import { Button } from "@/components/ui/button"
import type { ListResponse, ProductDTO } from "@/types"

export function ProductGrid({
  products,
  onConfigure,
}: {
  products: ListResponse<ProductDTO>
  onConfigure: (p: ProductDTO) => void
}) {
  const [activeCat, setActiveCat] = useState<string>("all")

  const getCatPos = (c?: { position?: number | null } | null) => {
    const v = c?.position
    return typeof v === "number" && isFinite(v) ? v : 1_000_000
  }

  const activeItems = products.items.filter((it) => it.isActive !== false)

  const sortedProducts = [...activeItems].sort((a, b) => {
    const pa = getCatPos(a.category)
    const pb = getCatPos(b.category)
    if (pa !== pb) return pa - pb
    return a.name.localeCompare(b.name)
  })

  const cats = useMemo(() => {
    const byId = new Map<string, { id: string; name: string; position: number }>()
    for (const p of activeItems) {
      const c = p.category
      if (c?.id) {
        byId.set(c.id, { id: c.id, name: c.name, position: getCatPos(c) })
      }
    }
    return Array.from(byId.values()).sort((a, b) => a.position - b.position || a.name.localeCompare(b.name))
  }, [activeItems])

  const countsByCatId = useMemo(() => {
    const counts: Record<string, number> = {}
    for (const it of activeItems) {
      const id = it.category?.id
      if (!id) continue
      counts[id] = (counts[id] || 0) + 1
    }
    return counts
  }, [activeItems])

  const filtered = activeCat === "all"
    ? sortedProducts
    : sortedProducts.filter((it) => it.category?.id === activeCat)

  return (
    <main className="container mx-auto h-full space-y-3 overflow-y-auto px-3 pb-4 md:px-4">
      <h1 className="mb-1 text-xl font-semibold">Menu</h1>
      <div className="flex flex-wrap items-center gap-2">
        <Button
          className={`border-border hover:bg-card hover:text-foreground group flex h-10 items-center justify-between gap-2 rounded-[10px] border px-1.5 text-sm ${
            activeCat === "all" ? "bg-card" : "text-muted-foreground bg-transparent"
          }`}
          onClick={() => setActiveCat("all")}
          aria-pressed={activeCat === "all"}
        >
          Alles
          <span
            className={`text-foreground group-hover:text-foreground flex h-7 min-w-7 items-center justify-center rounded-[6px] px-1 text-sm transition-all duration-300 group-hover:bg-[#FFBBBB] ${
              activeCat === "all" ? "text-foreground bg-[#FFBBBB]" : "text-muted-foreground bg-[#D9D9D9]"
            }`}
          >
            {activeItems.length}
          </span>
        </Button>
        {cats.map((c) => (
          <Button
            key={c.id}
            className={`border-border hover:bg-card hover:text-foreground group flex h-10 items-center justify-between gap-2 rounded-[10px] border px-1.5 text-sm ${
              activeCat === c.id ? "bg-card" : "text-muted-foreground bg-transparent"
            }`}
            onClick={() => setActiveCat(c.id)}
            aria-pressed={activeCat === c.id}
          >
            {c.name}
            <span
              className={`text-foreground group-hover:text-foreground flex h-7 min-w-7 items-center justify-center rounded-[6px] px-1 text-sm transition-all duration-300 group-hover:bg-[#FFBBBB] ${
                activeCat === c.id ? "text-foreground bg-[#FFBBBB]" : "text-muted-foreground bg-[#D9D9D9]"
              }`}
            >
              {countsByCatId[c.id] ?? 0}
            </span>
          </Button>
        ))}
      </div>

      <div className="grid grid-cols-2 gap-3 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
        {filtered.map((product) => (
          <ProductCardPOS key={product.id} product={product} onConfigure={() => onConfigure(product)} />
        ))}
      </div>
    </main>
  )
}
