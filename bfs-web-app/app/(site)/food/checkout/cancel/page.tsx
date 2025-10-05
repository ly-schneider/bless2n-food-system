"use client"

import { X } from "lucide-react"
import { useRouter } from "next/navigation"
import { Button } from "@/components/ui/button"
import { useCart } from "@/contexts/cart-context"

export default function CheckoutCancelPage() {
  const router = useRouter()
  useCart() // keep context warm; not directly used

  const handleRetry = () => { router.push("/food/checkout") }

  return (
    <div className="flex min-h-[70vh] flex-col items-center justify-center gap-10 px-4 pb-36">
      <div className="relative h-72 w-72">
        {/* Concentric rings (error red gradient) */}
        <div className="absolute inset-0 rounded-full bg-gradient-to-br from-red-200/60 to-red-400/40 blur-sm" />
        <div className="absolute inset-8 rounded-full border-8 border-red-300/60" />
        <div className="absolute inset-16 rounded-full border-8 border-red-400/50" />
        <div className="absolute inset-24 flex items-center justify-center rounded-full bg-red-500/80">
          <X className="h-14 w-14 text-white" />
        </div>
      </div>
      <h1 className="text-3xl font-semibold">Bezahlung Abgebrochen</h1>

      {/* Bottom fixed action buttons stacked */}
      <div className="max-w-xl mx-auto fixed inset-x-0 bottom-0 p-4">
        <div className="flex flex-col gap-3">
          <Button className="rounded-pill h-12 w-full text-base font-medium" onClick={handleRetry}>
            Erneut versuchen
          </Button>
          <Button
            variant="outline"
            className="rounded-pill h-12 w-full bg-white text-base font-medium text-black"
            onClick={() => router.push("/food/checkout?from=cancel")}
          >
            Zurück zum Warenkorb
          </Button>
          {/* No dynamic errors in this view */}
        </div>
      </div>
    </div>
  )
}
