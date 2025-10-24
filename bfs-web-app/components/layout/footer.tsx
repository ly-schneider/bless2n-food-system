"use client"

import Link from "next/link"
import { usePathname } from "next/navigation"
import { useCart } from "@/contexts/cart-context"

export default function Footer() {
  const pathname = usePathname()
  const { cart } = useCart()
  const totalItems = cart.items.reduce((sum, it) => sum + it.quantity, 0)

  if (pathname?.startsWith("/food/checkout")) return null

  return (
    <footer
      id="app-footer"
      className={`text-muted-foreground w-full border-t border-gray-200/70 py-4 text-sm ${
        (pathname === "/" || pathname === "/food") && totalItems > 0
          ? "mb-20"
          : pathname === "/food/orders"
            ? "mb-18"
            : "mb-4"
      }`}
    >
      <div className="container mx-auto px-4">
        <nav className="flex flex-wrap items-center justify-center gap-x-4 gap-y-2">
          <Link href="/agb" className="hover:underline">
            AGB
          </Link>
          <span className="text-gray-300">•</span>
          <Link href="/datenschutz" className="hover:underline">
            Datenschutzerklärung
          </Link>
          <span className="text-gray-300">•</span>
          <Link
            href="https://github.com/ly-schneider/bless2n-food-system"
            className="hover:underline"
            target="_blank"
            rel="noopener noreferrer"
          >
            GitHub
          </Link>
        </nav>
      </div>
    </footer>
  )
}
