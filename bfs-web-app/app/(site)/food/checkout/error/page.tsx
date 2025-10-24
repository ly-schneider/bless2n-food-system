"use client"

import { AlertTriangle } from "lucide-react"
import { useRouter } from "next/navigation"
import { Button } from "@/components/ui/button"
import { useCart } from "@/contexts/cart-context"

export default function CheckoutErrorPage() {
  const router = useRouter()
  useCart()

  const handleRetry = () => {
    router.push("/food/checkout")
  }

  return (
    <div className="flex min-h-[70vh] flex-col items-center justify-center gap-10 px-4 pb-36">
      <div className="relative h-64 w-64 sm:h-72 sm:w-72">
        <div className="absolute inset-0 rounded-full bg-gradient-to-br from-red-200/60 to-red-400/40 blur-sm" />
        <div className="absolute inset-8 rounded-full border-8 border-red-300/60" />
        <div className="absolute inset-16 rounded-full border-8 border-red-400/50" />
        <div className="absolute inset-24 flex items-center justify-center rounded-full bg-red-500/80">
          <AlertTriangle className="h-9 w-9 text-white sm:h-14 sm:w-14" />
        </div>
      </div>
      <h1 className="text-center text-2xl font-semibold sm:text-3xl">Bezahlung fehlgeschlagen</h1>
      <p className="text-muted-foreground -mt-6 text-center">
        Bitte versuche es erneut oder wähle eine andere Methode.
      </p>

      <div className="fixed inset-x-0 bottom-0 mx-auto max-w-xl p-4">
        <div className="flex flex-col gap-3">
          <Button className="rounded-pill h-12 w-full text-base font-medium" onClick={handleRetry}>
            Erneut versuchen
          </Button>
          <Button
            variant="outline"
            className="rounded-pill h-12 w-full bg-white text-base font-medium text-black"
            onClick={() => router.push("/food/checkout?from=error")}
          >
            Zurück zum Warenkorb
          </Button>
        </div>
      </div>
    </div>
  )
}
