import type { Metadata } from "next"

export const metadata: Metadata = {
  title: "Zahlung wird bestätigt - BlessThun Food",
  description: "Wir prüfen den Zahlungsstatus und leiten dich weiter.",
  alternates: { canonical: "/food/checkout/payment/next" },
  openGraph: {
    title: "Zahlung wird bestätigt | BlessThun Food",
    description: "Bitte einen Moment Geduld – wir prüfen den Status.",
    url: "/food/checkout/payment/next",
    type: "website",
  },
}

export default function PaymentNextLayout({ children }: { children: React.ReactNode }) {
  return children
}
