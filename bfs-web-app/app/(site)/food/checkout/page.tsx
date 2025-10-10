import { Suspense } from "react"
import CheckoutClient from "./checkout-client"

export default function CheckoutPage() {
  return (
    <Suspense fallback={<div className="container mx-auto px-4 pt-4">Lade Warenkorb…</div>}>
      <CheckoutClient />
    </Suspense>
  )
}
