import type { Metadata } from "next"

export const metadata: Metadata = {
  title: "Bezahlung fehlgeschlagen - Bless2n Food System",
  description: "Die Zahlung ist fehlgeschlagen. Bitte versuche es erneut oder kehre zum Warenkorb zurück.",
  alternates: { canonical: "/checkout/error" },
  openGraph: {
    title: "Bezahlung fehlgeschlagen | Bless2n Food System",
    description: "Zahlung fehlgeschlagen – zurück zum Warenkorb oder erneut versuchen.",
    url: "/checkout/error",
    type: "website",
  },
}

export default function CheckoutErrorLayout({ children }: { children: React.ReactNode }) {
  return children
}

