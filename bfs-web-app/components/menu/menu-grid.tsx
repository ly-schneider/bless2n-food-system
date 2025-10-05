"use client"

import { Info } from "lucide-react"
import Image from "next/image"
import { useMemo, useState } from "react"
import { CartButtons } from "@/components/cart/cart-buttons"
import { ProductConfigurationModal } from "@/components/cart/product-configuration-modal"
import { Card, CardContent, CardHeader } from "@/components/ui/card"
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip"
import { formatChf } from "@/lib/utils"
import { ListResponse, ProductDTO } from "@/types"

export function MenuGrid({ products }: { products: ListResponse<ProductDTO> }) {
  const getCatPos = (p?: { position?: number | null } | null) => {
    const v = p?.position
    return typeof v === 'number' && isFinite(v) ? v : 1_000_000
  }

  const sortedProducts = [...products.items].sort((a, b) => {
    const pa = getCatPos(a.category)
    const pb = getCatPos(b.category)
    if (pa !== pb) return pa - pb
    return a.name.localeCompare(b.name)
  })

  return (
    <div className="grid gap-3 md:gap-5 grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
      {sortedProducts.map((product) => (
        <MenuProductCard key={product.id} product={product} />
      ))}
    </div>
  )
}

function MenuProductCard({ product }: { product: ProductDTO }) {
  const [isConfigModalOpen, setIsConfigModalOpen] = useState(false);
  const isAvailable = product.isAvailable !== false; // default true
  const isLowStock = product.isLowStock === true;
  const availableQty = product.availableQuantity ?? null;
  const isActive = product.isActive !== false;
  const composition = useMemo(() => {
    if (product.type !== 'menu' || !product.menu?.slots || product.menu.slots.length === 0) return null
    const counts = new Map<string, number>()
    for (const slot of product.menu.slots) {
      const name = slot.name?.trim() || 'Slot'
      counts.set(name, (counts.get(name) || 0) + 1)
    }
    // Build minimal description like: "Burger + Beilage + 2× Getränk"
    const parts = Array.from(counts.entries()).map(([name, count]) => (count > 1 ? `${count}× ${name}` : name))
    return parts.join(' + ')
  }, [product])
  
  const handleConfigureProduct = () => {
    setIsConfigModalOpen(true);
  };
  
  return (
    <>
      <Card className="overflow-hidden transition-shadow hover:shadow-lg p-0 rounded-[11px] gap-0">
        <CardHeader className="p-2">
          <div className="relative aspect-video rounded-t-lg bg-[#cec9c6] rounded-[11px]">
            {product.image ? (
              <Image
                src={product.image}
                alt={"Produktbild von " + product.name}
                fill
                sizes="(max-width: 640px) 50vw, (max-width: 1024px) 33vw, 25vw"
                quality={90}
                className="w-full h-full object-cover rounded-[11px]"
              />
            ) : (
              <div className="absolute inset-0 flex items-center justify-center text-zinc-500">
                Kein Bild
              </div>
            )}
            {(!isAvailable) && (
              <div className="absolute inset-0 z-10 grid place-items-center bg-black/55 rounded-[11px]">
                <span className="px-3 py-1 text-sm font-medium text-white bg-red-400 rounded-full">Ausverkauft</span>
              </div>
            )}
            {isAvailable && !isActive && (
              <div className="absolute inset-0 z-10 grid place-items-center bg-black/55 rounded-[11px]">
                <span className="px-3 py-1 text-sm font-medium text-white bg-zinc-700 rounded-full">Nicht verfügbar</span>
              </div>
            )}
            {isLowStock && isAvailable && isActive && (
              <div className="absolute top-1 left-2 z-10">
                <span className="px-2 py-0.5 text-xs font-medium text-white bg-amber-600 rounded-full">
                  {availableQty !== null ? `Nur ${availableQty} übrig` : 'Geringer Bestand'}
                </span>
              </div>
            )}
            {composition && (
              <Tooltip>
                <TooltipTrigger asChild>
                  <button
                    type="button"
                    aria-label="Menüinhalt anzeigen"
                    className="absolute top-1 right-1 z-10 inline-flex size-7 items-center justify-center rounded-full bg-white text-black shadow-sm ring-1 ring-black/10 hover:bg-white/90"
                  >
                    <Info className="size-4" />
                  </button>
                </TooltipTrigger>
                <TooltipContent side="left" align="start" sideOffset={6} className="bg-white text-black rounded-full px-3 py-1">
                  {composition}
                </TooltipContent>
              </Tooltip>
            )}
          </div>
        </CardHeader>

        <CardContent className="px-2 pt-0 pb-4">
          <div className="flex items-center justify-between">
            <div className="flex flex-col">
              <h3 className="text-lg font-family-secondary">{product.name}</h3>
              <p className="text-base font-family-secondary">{formatChf(product.priceCents)}</p>
            </div>
            <div className="flex items-center">
              <CartButtons 
                product={product}
                onConfigureProduct={handleConfigureProduct}
                disabled={!isAvailable || !isActive}
              />
            </div>
          </div>
        </CardContent>
      </Card>
      
      <ProductConfigurationModal
        product={product}
        isOpen={isConfigModalOpen}
        onClose={() => setIsConfigModalOpen(false)}
      />
    </>
  )
}

export default MenuGrid
