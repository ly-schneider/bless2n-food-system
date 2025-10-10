import { Suspense } from "react"
import PaymentNextClient from "./payment-next-client"

export default function PaymentNextPage() {
  return (
    <Suspense
      fallback={
        <div className="flex min-h-[70vh] items-center justify-center">
          <p className="text-muted-foreground">Zahlungsstatus wird geprüft…</p>
        </div>
      }
    >
      <PaymentNextClient />
    </Suspense>
  )
}
