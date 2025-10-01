import type { Metadata } from "next"

export const metadata: Metadata = {
  title: "Zahlung - BlessThun Food",
  description: "Mit TWINT via Stripe Payment Element bezahlen.",
  alternates: { canonical: "/checkout/payment" },
  openGraph: {
    title: "Zahlung | BlessThun Food",
    description: "TWINT Zahlung über Stripe Payment Element starten.",
    url: "/checkout/payment",
    type: "website",
  },
}

export default function PaymentLayout({ children }: { children: React.ReactNode }) {
  return children
}
