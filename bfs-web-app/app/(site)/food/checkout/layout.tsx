import type { Metadata } from "next"

export const metadata: Metadata = {
  title: "Warenkorb",
  description: "Prüfe deinen Warenkorb und starte die Zahlung bei BlessThun Food.",
  alternates: { canonical: "/food/checkout" },
  openGraph: {
    title: "Warenkorb",
    description: "Warenkorb prüfen und zur Zahlung fortfahren.",
    url: "/food/checkout",
  },
}

export default function CheckoutLayout({ children }: { children: React.ReactNode }) {
  return <div className="mx-auto max-w-xl">{children}</div>
}
