import { Metadata } from "next"
import Header from "@/components/layout/header"
import MenuGrid from "@/components/menu/menu-grid"
import ProductAPI from "@/lib/api/products"
import { ListProductsResponse } from "@/types"

export const metadata: Metadata = {
  title: "Menu - Bless2n Food System",
  description: "Explore our delicious menu of fresh, carefully prepared meals.",
}

export default async function HomePage() {
  // Fetch products directly via products API
  let products: ListProductsResponse
  try {
    const res = await ProductAPI.listPublicProducts({ activeOnly: true, limit: 100 })
    products = res
  } catch {
    products = { products: [], items: [], total: 0 }
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
