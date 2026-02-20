import type { Metadata } from "next"

export const metadata: Metadata = {
  title: "Zahlung",
  description: "Bezahle deine Bestellung bequem mit TWINT bei BlessThun Food.",
  alternates: { canonical: "/food/checkout/payment" },
  openGraph: {
    title: "Zahlung",
    description: "TWINT-Zahlung für deine BlessThun Food Bestellung starten.",
    url: "/food/checkout/payment",
  },
}

export default function PaymentLayout({ children }: { children: React.ReactNode }) {
  return children
}
