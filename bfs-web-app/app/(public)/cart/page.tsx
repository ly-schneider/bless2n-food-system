import { ArrowRight, ShoppingCart } from "lucide-react"
import { Metadata } from "next"
import Link from "next/link"
import { Button } from "@/components/ui/button"

export const metadata: Metadata = {
  title: "Your Cart - Bless2n Food System",
  description: "Review your order and proceed to checkout.",
}

export default function CartPage() {
  // Real cart integration pending backend endpoints/context.
  // For now, show empty cart without mock data.
  return (
    <div className="container mx-auto px-4 py-16">
      <div className="space-y-6 text-center">
        <div className="bg-muted mx-auto flex h-24 w-24 items-center justify-center rounded-full">
          <ShoppingCart className="text-muted-foreground h-12 w-12" />
        </div>

        <div className="space-y-2">
          <h1 className="text-2xl font-bold">Your cart is empty</h1>
          <p className="text-muted-foreground">Looks like you haven't added any items to your cart yet.</p>
        </div>

        <Button size="lg" asChild>
          <Link href="/menu">
            Start Shopping
            <ArrowRight className="ml-2 h-4 w-4" />
          </Link>
        </Button>
      </div>
    </div>
  )
}
