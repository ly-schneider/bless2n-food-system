import type { Metadata } from "next"

export const metadata: Metadata = {
  title: "Zahlung - BlessThun Food",
  description: "Mit TWINT bezahlen.",
  alternates: { canonical: "/food/checkout/payment" },
  openGraph: {
    title: "Zahlung | BlessThun Food",
    description: "TWINT Zahlung starten.",
    url: "/food/checkout/payment",
    type: "website",
  },
}

export default function PaymentLayout({ children }: { children: React.ReactNode }) {
  return children
}
