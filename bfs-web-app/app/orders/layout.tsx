import type { Metadata } from "next"

export const metadata: Metadata = {
  title: "Bestellungen - Bless2n Food System",
  description: "Übersicht Ihrer Bestellungen und QR-Codes zur Abholung.",
  alternates: { canonical: "/orders" },
  openGraph: {
    title: "Bestellungen | Bless2n Food System",
    description: "Verwalten Sie Ihre Bestellungen und Abholungs-QR-Codes.",
    url: "/orders",
    type: "website",
  },
}

export default function OrdersLayout({ children }: { children: React.ReactNode }) {
  return children
}

