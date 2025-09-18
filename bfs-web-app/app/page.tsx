import { Metadata } from "next"
import Header from "@/components/layout/header"
import MenuGrid from "@/components/menu/menu-grid"
import { listProducts } from "@/lib/api/products"
import { ListResponse, ProductDTO } from "@/types"

export const metadata: Metadata = {
  title: "Menu - Bless2n Food System",
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
      <Header />

      <main className="container mx-auto px-4 py-8">
        <h2 className="text-2xl mb-2">Alle Produkte</h2>
        <MenuGrid products={products} />
      </main>
    </div>
  )
}
