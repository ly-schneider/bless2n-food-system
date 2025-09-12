"use client"
import { Plus } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardFooter, CardHeader } from "@/components/ui/card"

export type MenuItem = {
  id: string
  name: string
  description?: string
  price: number
  isActive: boolean
}

export function MenuGrid({ items }: { items: MenuItem[] }) {
  return (
    <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
      {items.map((item) => (
        <MenuItemCard key={item.id} item={item} />
      ))}
    </div>
  )
}

function MenuItemCard({ item }: { item: MenuItem }) {
  return (
    <Card className="overflow-hidden transition-shadow hover:shadow-lg">
      <CardHeader className="p-0">
        <div className="bg-muted relative aspect-video rounded-t-lg">
          <div className="text-muted-foreground absolute inset-0 flex items-center justify-center">No Image</div>
          {!item.isActive && (
            <div className="absolute inset-0 flex items-center justify-center bg-black/50">
              <span className="font-semibold text-white">Currently Unavailable</span>
            </div>
          )}
        </div>
      </CardHeader>

      <CardContent className="p-4">
        <div className="mb-2 flex items-start justify-between">
          <h3 className="line-clamp-1 text-lg font-semibold">{item.name}</h3>
          <span className="text-primary font-bold">${item.price}</span>
        </div>
        {item.description && (
          <p className="text-muted-foreground mb-3 line-clamp-2 text-sm">{item.description}</p>
        )}
      </CardContent>

      <CardFooter className="p-4 pt-0">
        <AddToCartButton item={item} disabled={!item.isActive} />
      </CardFooter>
    </Card>
  )
}

function AddToCartButton({ item, disabled }: { item: Pick<MenuItem, "id" | "name" | "price">; disabled?: boolean }) {
  const handleAddToCart = () => {
    // TODO: Hook up to backend cart/order draft when available
    console.log("Add to cart clicked:", item.id)
  }

  return (
    <Button onClick={handleAddToCart} disabled={disabled} className="w-full">
      <Plus className="mr-2 h-4 w-4" />
      {disabled ? "Unavailable" : "Add to Cart"}
    </Button>
  )
}

export default MenuGrid

