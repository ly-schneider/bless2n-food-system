"use client"

import { useRouter, useSearchParams } from "next/navigation"
import { ArrowLeft } from "lucide-react"
import QRCode from "@/components/qrcode"
import { Button } from "@/components/ui/button"

export default function CheckoutQRPage() {
  const sp = useSearchParams()
  const orderId = sp.get("order_id")
  const router = useRouter()

  return (
    <div className="container mx-auto px-4 py-16 pb-28 flex flex-col">
      <h1 className="text-2xl font-semibold mb-2">Abholungs QR-Code</h1>
      <p className="mb-12 text-muted-foreground">Zeigen Sie diesen QR-Code bei der Abholung vor.</p>
      {orderId ? (
        <QRCode value={orderId} size={260} className="rounded-[11px] border-2 mx-auto p-1" />
      ) : (
        <p className="text-red-600">Bestellnummer fehlt.</p>
      )}

      <div className="fixed inset-x-0 bottom-0 z-50 p-4">
        <Button
          variant="outline"
          className="rounded-pill w-full h-12 text-base font-medium bg-white text-black"
          onClick={() => router.push("/")}
        >
          <ArrowLeft style={{ width: 20, height: 20 }} /> Zum Menü
        </Button>
      </div>
    </div>
  )
}
