import type { Metadata } from "next"

export const metadata: Metadata = {
  title: "Zahlung - Bless2n Food System",
  description: "Mit TWINT via Stripe Checkout bezahlen.",
  alternates: { canonical: "/checkout/payment" },
  openGraph: {
    title: "Zahlung | Bless2n Food System",
    description: "TWINT Zahlung über Stripe Checkout starten.",
    url: "/checkout/payment",
    type: "website",
  },
}

export default function PaymentLayout({ children }: { children: React.ReactNode }) {
  return children
}

