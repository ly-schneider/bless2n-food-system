import type { Metadata } from "next"

export const metadata: Metadata = {
  title: "Abholungs QR-Code - Bless2n Food System",
  description: "QR-Code für die Abholung Ihrer Bestellung anzeigen.",
  alternates: { canonical: "/checkout/qr" },
  openGraph: {
    title: "Abholungs QR-Code | Bless2n Food System",
    description: "Zeigen Sie den QR-Code bei der Abholung vor.",
    url: "/checkout/qr",
    type: "website",
  },
}

export default function CheckoutQRLayout({ children }: { children: React.ReactNode }) {
  return children
}

