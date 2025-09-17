import { Metadata } from "next"

import MenuGrid from "@/components/menu/menu-grid"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import ProductAPI from "@/lib/api/products"
import { ListProductsResponse } from "@/types"

export const metadata: Metadata = {
  title: "Menu - Bless2n Food System",
  description: "Explore our delicious menu of fresh, carefully prepared meals.",
}

export default async function MenuPage() {
  // Fetch products directly via products API
  let products: ListProductsResponse = { products: [], items: [], total: 0 }
  try {
    const res = await ProductAPI.listPublicProducts({ activeOnly: true, limit: 100 })
    products = res
  } catch {
    products = { products: [], items: [], total: 0 }
  }

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="mb-8">
        <h1 className="mb-4 text-3xl font-bold">Our Menu</h1>
        <p className="text-muted-foreground">
          Fresh, delicious meals made with the finest ingredients. All dishes are prepared to order.
        </p>
      </div>

      <Tabs defaultValue="all" className="w-full">
        <TabsList className="grid w-full grid-cols-1">
          <TabsTrigger value="all">All Items</TabsTrigger>
        </TabsList>

        <TabsContent value="all" className="mt-6">
          <MenuGrid products={products} />
        </TabsContent>
      </Tabs>
    </div>
  )
}
