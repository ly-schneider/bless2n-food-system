import type { Metadata } from "next"

export async function generateMetadata(
  { params }: { params: Promise<{ orderId: string }> }
): Promise<Metadata> {
  const { orderId } = await params
  const canonical = `/food/orders/${encodeURIComponent(orderId)}`
  return {
    title: "Abholungs QR-Code - BlessThun Food",
    description: "QR-Code für die Abholung Ihrer Bestellung anzeigen.",
    alternates: { canonical },
    openGraph: {
      title: "Abholungs QR-Code | BlessThun Food",
      description: "Zeigen Sie den QR-Code bei der Abholung vor.",
      url: canonical,
      type: "website",
    },
  }
}

export default function CheckoutQRLayout({ children }: { children: React.ReactNode }) {
  return children
}
