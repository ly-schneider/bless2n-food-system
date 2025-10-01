import type { Metadata } from "next"

export const metadata: Metadata = {
  title: "Bezahlung abgebrochen - BlessThun Food",
  description: "Die Zahlung wurde abgebrochen. Versuchen Sie es erneut oder kehren Sie zum Warenkorb zurück.",
  alternates: { canonical: "/checkout/cancel" },
  openGraph: {
    title: "Bezahlung abgebrochen | BlessThun Food",
    description: "Zahlung nicht abgeschlossen – zurück zum Warenkorb oder erneut versuchen.",
    url: "/checkout/cancel",
    type: "website",
  },
}

export default function CheckoutCancelLayout({ children }: { children: React.ReactNode }) {
  return children
}

