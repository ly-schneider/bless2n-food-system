import type { Metadata } from "next"

export const metadata: Metadata = {
  title: "Warenkorb - Bless2n Food System",
  description: "Prüfen Sie Ihren Warenkorb und starten Sie die Zahlung.",
  alternates: { canonical: "/checkout" },
  openGraph: {
    title: "Warenkorb | Bless2n Food System",
    description: "Warenkorb prüfen und zur Zahlung fortfahren.",
    url: "/checkout",
    type: "website",
  },
}

export default function CheckoutLayout({ children }: { children: React.ReactNode }) {
  return children
}

