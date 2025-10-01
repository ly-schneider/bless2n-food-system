import type { Metadata } from "next"

export const metadata: Metadata = {
  title: "Zahlung wird bestätigt - Bless2n Food System",
  description: "Wir prüfen den Zahlungsstatus und leiten dich weiter.",
  alternates: { canonical: "/checkout/payment/next" },
  openGraph: {
    title: "Zahlung wird bestätigt | Bless2n Food System",
    description: "Bitte einen Moment Geduld – wir prüfen den Status.",
    url: "/checkout/payment/next",
    type: "website",
  },
}

export default function PaymentNextLayout({ children }: { children: React.ReactNode }) {
  return children
}

