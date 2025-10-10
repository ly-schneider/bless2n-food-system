import { Suspense } from "react"
import CheckoutSuccessClient from "./success-client"

export default function CheckoutSuccessPage() {
  return (
    <Suspense fallback={<div className="flex min-h-[70vh] items-center justify-center">Lade Erfolg…</div>}>
      <CheckoutSuccessClient />
    </Suspense>
  )
}
