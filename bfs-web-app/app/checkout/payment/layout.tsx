import type { Metadata } from "next"

export const metadata: Metadata = {
  title: "Zahlung - Bless2n Food System",
  description: "Mit TWINT via Stripe Payment Element bezahlen.",
  alternates: { canonical: "/checkout/payment" },
  openGraph: {
    title: "Zahlung | Bless2n Food System",
    description: "TWINT Zahlung über Stripe Payment Element starten.",
    url: "/checkout/payment",
    type: "website",
  },
}

export default function PaymentLayout({ children }: { children: React.ReactNode }) {
  return children
}
