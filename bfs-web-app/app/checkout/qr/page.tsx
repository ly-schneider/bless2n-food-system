"use client"

import { ArrowLeft } from "lucide-react"
import { useRouter, useSearchParams } from "next/navigation"
import { useEffect, useRef } from "react"
import QRCode from "@/components/qrcode"
import { Button } from "@/components/ui/button"
import { useCart } from "@/contexts/cart-context"
import { addOrder } from "@/lib/orders-storage"

export default function CheckoutQRPage() {
  const sp = useSearchParams()
  const orderId = sp.get("order_id")
  const from = sp.get("from")
  const { clearCart } = useCart()
  const clearedRef = useRef(false)

  useEffect(() => {
    if (orderId) addOrder(orderId)
  }, [orderId])

  // Ensure cart is cleared when landing on QR page as a fallback
  useEffect(() => {
    if (!clearedRef.current) {
      clearedRef.current = true
      clearCart()
    }
  }, [clearCart])
  const router = useRouter()

  return (
    <div className="p-4 pb-28 flex flex-col">
      <h1 className="text-2xl font-semibold mb-2">Abholungs QR-Code</h1>
      <p className={`text-muted-foreground text-sm ${from === "success" ? "mb-2" : "mb-8"}`}>Zeigen Sie diesen QR-Code bei der Abholung vor.</p>
      {from === "success" && <p className="mb-8 text-muted-foreground text-sm">Du kannst diesen QR-Code jederzeit in deinen Bestellungen finden.</p>}
      {orderId ? (
        <QRCode value={orderId} size={260} className="rounded-[11px] border-2 mx-auto p-1" />
      ) : (
        <p className="text-red-600">Bestellnummer fehlt.</p>
      )}

      <div className="fixed inset-x-0 bottom-0 z-50 p-4">
        <div className="flex flex-col gap-2">
          <Button
            variant="outline"
            className="rounded-pill w-full h-12 text-base font-medium bg-white text-black"
            onClick={() => router.push(from === "orders" ? "/orders" : "/")}
          >
            <ArrowLeft style={{ width: 20, height: 20 }} /> {from === "orders" ? "Zurück" : "Zum Menü"}
          </Button>

          {from === "success" && (
            <Button
              variant="selected"
              className="rounded-pill w-full h-12 text-base font-medium"
              onClick={() => router.push("/orders")}
            >
              Alle Bestellungen
            </Button>
          )}
        </div>
      </div>
    </div>
  )
}
