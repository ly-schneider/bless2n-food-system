import { Metadata } from "next"
import { FloatingBottomNav } from "@/components/cart/floating-bottom-nav"
import MenuGrid from "@/components/menu/menu-grid"
import { listProducts } from "@/lib/api/products"
import { ListResponse, ProductDTO } from "@/types"

export const metadata: Metadata = {
  title: "Menu - BlessThun Food",
  description: "Explore our delicious menu of fresh, carefully prepared meals.",
}

export default async function HomePage() {
  let products: ListResponse<ProductDTO>
  try {
    products = await listProducts()
  } catch (error) {
    console.error('Failed to fetch products:', error)
    products = { items: [], count: 0 }
  }

  return (
    <div className="bg-background min-h-screen">

      <main className="container mx-auto p-4">
        <h2 className="text-2xl mb-2">Alle Produkte</h2>
        <MenuGrid products={products} />
      </main>

      <FloatingBottomNav />
    </div>
  )
}
