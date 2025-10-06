import type { Metadata } from "next"
import { Golos_Text } from "next/font/google"
import Header from "@/components/layout/header"
import Footer from "@/components/layout/footer"
import { CartProvider } from "@/contexts/cart-context"

const golosText = Golos_Text({
  weight: ["500"],
  subsets: ["latin"],
  variable: "--font-golos-text",
})

export const metadata: Metadata = {
  other: {
    "theme-color": "#E9E7E6",
    "msapplication-TileColor": "#E9E7E6",
    "apple-mobile-web-app-status-bar-style": "default",
  },
}

export default function SiteLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className={`${golosText.variable} min-h-screen overflow-x-hidden flex flex-col`}>
      <CartProvider>
        <Header />
        <main className="flex-1">{children}</main>
        <Footer />
      </CartProvider>
    </div>
  )
}
