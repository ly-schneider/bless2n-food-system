import type { Metadata } from "next"

export const metadata: Metadata = {
  title: "Zahlung wird bestätigt",
  description: "Dein Zahlungsstatus wird geprüft – du wirst gleich weitergeleitet.",
  alternates: { canonical: "/food/checkout/payment/next" },
  openGraph: {
    title: "Zahlung wird bestätigt",
    description: "Bitte einen Moment Geduld – wir prüfen den Status deiner Zahlung.",
    url: "/food/checkout/payment/next",
  },
}

export default function PaymentNextLayout({ children }: { children: React.ReactNode }) {
  return children
}
