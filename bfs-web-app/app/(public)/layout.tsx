import { Menu, ShoppingCart, User } from "lucide-react"
import Link from "next/link"
import { Suspense } from "react"
import { Button } from "@/components/ui/button"

interface PublicLayoutProps {
  children: React.ReactNode
}

export default function PublicLayout({ children }: PublicLayoutProps) {
  return (
    <div className="bg-background min-h-screen">
      <header className="bg-background/95 supports-[backdrop-filter]:bg-background/60 sticky top-0 z-50 border-b backdrop-blur">
        <div className="container mx-auto flex h-16 items-center justify-between px-4">
          <Link href="/" className="text-xl font-bold">
            Bless2n Food
          </Link>

          <nav className="hidden items-center space-x-6 md:flex">
            <Link href="/" className="hover:text-primary text-sm font-medium transition-colors">
              Menu
            </Link>
            <Link href="/about" className="hover:text-primary text-sm font-medium transition-colors">
              About
            </Link>
            <Link href="/contact" className="hover:text-primary text-sm font-medium transition-colors">
              Contact
            </Link>
          </nav>

          <div className="flex items-center space-x-4">
            <Suspense fallback={<div className="h-8 w-8" />}>
              <CartButton />
            </Suspense>

            <Button variant="ghost" size="sm" asChild>
              <Link href="/login">
                <User className="mr-2 h-4 w-4" />
                Login
              </Link>
            </Button>

            <Button variant="ghost" size="sm" className="md:hidden">
              <Menu className="h-4 w-4" />
            </Button>
          </div>
        </div>
      </header>

      <main className="flex-1">{children}</main>

      <footer className="bg-muted/50 border-t">
        <div className="container mx-auto px-4 py-8">
          <div className="grid grid-cols-1 gap-8 md:grid-cols-4">
            <div>
              <h3 className="mb-4 font-semibold">Bless2n Food System</h3>
              <p className="text-muted-foreground text-sm">
                Fresh, delicious meals crafted with care and delivered with love.
              </p>
            </div>

            <div>
              <h4 className="mb-4 font-semibold">Quick Links</h4>
              <ul className="space-y-2 text-sm">
                <li>
                  <Link href="/" className="text-muted-foreground hover:text-foreground">
                    Menu
                  </Link>
                </li>
                <li>
                  <Link href="/about" className="text-muted-foreground hover:text-foreground">
                    About Us
                  </Link>
                </li>
                <li>
                  <Link href="/contact" className="text-muted-foreground hover:text-foreground">
                    Contact
                  </Link>
                </li>
              </ul>
            </div>

            <div>
              <h4 className="mb-4 font-semibold">Support</h4>
              <ul className="space-y-2 text-sm">
                <li>
                  <Link href="/help" className="text-muted-foreground hover:text-foreground">
                    Help Center
                  </Link>
                </li>
                <li>
                  <Link href="/privacy" className="text-muted-foreground hover:text-foreground">
                    Privacy Policy
                  </Link>
                </li>
                <li>
                  <Link href="/terms" className="text-muted-foreground hover:text-foreground">
                    Terms of Service
                  </Link>
                </li>
              </ul>
            </div>

            <div>
              <h4 className="mb-4 font-semibold">Contact</h4>
              <ul className="text-muted-foreground space-y-2 text-sm">
                <li>Phone: (555) 123-4567</li>
                <li>Email: support@bless2nfood.com</li>
                <li>Hours: 9 AM - 10 PM daily</li>
              </ul>
            </div>
          </div>

          <div className="text-muted-foreground mt-8 border-t pt-8 text-center text-sm">
            © 2025 Bless2n Food System. All rights reserved.
          </div>
        </div>
      </footer>
    </div>
  )
}

// Cart button component with cart count
function CartButton() {
  // TODO: Connect to cart state management
  const itemCount = 0

  return (
    <Button variant="ghost" size="sm" asChild>
      <Link href="/cart" className="relative">
        <ShoppingCart className="mr-2 h-4 w-4" />
        Cart
        {itemCount > 0 && (
          <span className="bg-primary text-primary-foreground absolute -top-1 -right-1 flex h-5 w-5 items-center justify-center rounded-full text-xs">
            {itemCount}
          </span>
        )}
      </Link>
    </Button>
  )
}
