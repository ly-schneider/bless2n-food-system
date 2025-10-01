import type { Metadata } from "next"

export const metadata: Metadata = {
  title: "Bezahlung erfolgreich - BlessThun Food",
  description: "Ihre Zahlung war erfolgreich. Weiter zum Abholungs-QR-Code.",
  alternates: { canonical: "/checkout/success" },
  openGraph: {
    title: "Bezahlung erfolgreich | BlessThun Food",
    description: "Zahlung abgeschlossen – QR-Code zur Abholung anzeigen.",
    url: "/checkout/success",
    type: "website",
  },
}

export default function CheckoutSuccessLayout({ children }: { children: React.ReactNode }) {
  return children
}

