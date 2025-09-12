import { Metadata } from "next"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import ProductAPI from "@/lib/api/products"
import MenuGrid, { type MenuItem } from "@/components/menu/menu-grid"

export const metadata: Metadata = {
  title: "Menu - Bless2n Food System",
  description: "Explore our delicious menu of fresh, carefully prepared meals.",
}

export default async function MenuPage() {
  // Fetch products directly via products API
  let products: MenuItem[] = []
  try {
    const res = await ProductAPI.listPublicProducts({ activeOnly: true, limit: 100 })
    products = (res.products || []).map((p) => ({
      id: p.id,
      name: p.name,
      description: p.description,
      price: p.price,
      isActive: p.isActive,
    }))
  } catch {
    products = []
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
          <MenuGrid items={products} />
        </TabsContent>
      </Tabs>
    </div>
  )
}
