import { Metadata } from "next"
import { FloatingBottomNav } from "@/components/cart/floating-bottom-nav"
import { MenuGridLive } from "@/components/menu/menu-grid-live"
import { listProducts } from "@/lib/api/products"
import { getSystemStatus } from "@/lib/api/system"
import { ListResponse, ProductDTO } from "@/types"

export const metadata: Metadata = {
  title: "Menu",
  description: "Entdecke unser Angebot an frisch zubereiteten Gerichten und bestelle direkt online.",
  openGraph: {
    title: "Menu",
    description: "Frische Gerichte entdecken und online bestellen bei BlessThun.",
    url: "/food",
  },
}

export const dynamic = "force-dynamic"

export default async function HomePage() {
  const { enabled } = await getSystemStatus()

  if (!enabled) {
    return (
      <div className="bg-background flex min-h-screen items-center justify-center">
        <div className="text-center">
          <h2 className="text-2xl font-semibold">Aktuell geschlossen</h2>
          <p className="text-muted-foreground mt-2">Das Bestellsystem ist momentan nicht verfügbar.</p>
        </div>
      </div>
    )
  }

  let products: ListResponse<ProductDTO>
  try {
    products = await listProducts()
  } catch (error) {
    console.error("Failed to fetch products:", error)
    products = { items: [], count: 0 }
  }

  return (
    <div className="bg-background min-h-screen">
      <main className="container mx-auto p-4">
        <h2 className="mb-2 text-2xl">Alle Produkte</h2>
        <MenuGridLive initialProducts={products} />
      </main>

      <FloatingBottomNav />
    </div>
  )
}
