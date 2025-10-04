import type { Metadata } from "next"

export const metadata: Metadata = {
  title: "Warenkorb - BlessThun Food",
  description: "Prüfen Sie Ihren Warenkorb und starten Sie die Zahlung.",
  alternates: { canonical: "/food/checkout" },
  openGraph: {
    title: "Warenkorb | BlessThun Food",
    description: "Warenkorb prüfen und zur Zahlung fortfahren.",
    url: "/food/checkout",
    type: "website",
  },
}

export default function CheckoutLayout({ children }: { children: React.ReactNode }) {
  return children
}
