import type { Metadata } from "next"

export const metadata: Metadata = {
  title: "Bezahlung erfolgreich",
  description: "Deine Zahlung war erfolgreich – zeige den QR-Code bei der Abholung vor.",
  alternates: { canonical: "/food/checkout/success" },
  openGraph: {
    title: "Bezahlung erfolgreich",
    description: "Zahlung abgeschlossen – QR-Code zur Abholung anzeigen.",
    url: "/food/checkout/success",
  },
}

export default function CheckoutSuccessLayout({ children }: { children: React.ReactNode }) {
  return children
}
