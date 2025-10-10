import { Suspense } from "react"
import OrderPageClient from "./page-client"

export default function CheckoutQRPage() {
  return (
    <Suspense fallback={<div className="flex min-h-[70vh] items-center justify-center">Lade Bestellung…</div>}>
      <OrderPageClient />
    </Suspense>
  )
}
