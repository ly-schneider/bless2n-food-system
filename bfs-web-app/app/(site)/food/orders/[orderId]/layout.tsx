import type { Metadata } from "next"

export async function generateMetadata({ params }: { params: Promise<{ orderId: string }> }): Promise<Metadata> {
  const { orderId } = await params
  const canonical = `/food/orders/${encodeURIComponent(orderId)}`
  return {
    title: "Abholungs QR-Code",
    description: "Zeige den QR-Code bei der Abholung deiner BlessThun Food Bestellung vor.",
    alternates: { canonical },
    openGraph: {
      title: "Abholungs QR-Code",
      description: "QR-Code bei der Abholung vorzeigen.",
      url: canonical,
    },
  }
}

export default function CheckoutQRLayout({ children }: { children: React.ReactNode }) {
  return children
}
