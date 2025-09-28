"use client"

import { useState } from "react"
import { CartButtons } from "@/components/cart/cart-buttons"
import { ProductConfigurationModal } from "@/components/cart/product-configuration-modal"
import { Card, CardContent, CardHeader } from "@/components/ui/card"
import { formatChf } from "@/lib/utils"
import { ListResponse, ProductDTO } from "@/types"

export function MenuGrid({ products }: { products: ListResponse<ProductDTO> }) {
  const categoryOrder = {
    'Menus': 1,
    'Burgers': 2,
    'Beilagen': 3,
    'Getränke': 4
  }

  const sortedProducts = [...products.items].sort((a, b) => {
    const orderA = categoryOrder[a.category.name as keyof typeof categoryOrder] || 999
    const orderB = categoryOrder[b.category.name as keyof typeof categoryOrder] || 999
    return orderA - orderB
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
  
  const handleConfigureProduct = () => {
    setIsConfigModalOpen(true);
  };
  
  return (
    <>
      <Card className="overflow-hidden transition-shadow hover:shadow-lg p-0 rounded-[11px] gap-0">
        <CardHeader className="p-2">
          <div className="relative aspect-video rounded-t-lg bg-[#cec9c6] rounded-[11px]">
            {product.image ? (
              <img
                src={product.image}
                alt={"Produktbild von " + product.name}
                className="w-full h-full object-cover rounded-[11px]"
              />
            ) : (
              <div className="absolute inset-0 flex items-center justify-center text-zinc-500">
                Kein Bild
              </div>
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
