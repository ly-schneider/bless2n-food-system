"use client"

import { ArrowLeft, ArrowRight, LayoutList, QrCode } from "lucide-react"
import Link from "next/link"
import { useRouter } from "next/navigation"
import { useEffect, useState } from "react"
import { Button } from "@/components/ui/button"
import { getOrders, type StoredOrder } from "@/lib/orders-storage"

export default function OrdersPage() {
  const [orders, setOrders] = useState<StoredOrder[]>([])
  const router = useRouter()

  useEffect(() => {
    // TODO: If account session exists, fetch orders from backend here.
    setOrders(getOrders())
  }, [])

  return (
    <div className="p-4 pb-28">
      <h1 className="mb-2 text-2xl font-semibold">Bestellungen</h1>
      <p className="text-muted-foreground mb-6 text-sm">
        Tippe auf eine Bestellung, um den Abholungs QR-Code anzuzeigen.
      </p>

      {orders.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-12 text-center">
          <LayoutList className="text-muted-foreground mb-4 size-9" />
          <p className="text-muted-foreground text-lg font-semibold">Keine Bestellungen gefunden</p>
        </div>
      ) : (
        <ul className="flex flex-col gap-4 overflow-hidden">
          {orders.map((o) => {
            const dateLabel = new Date(o.createdAt).toLocaleDateString()
            return (
              <li key={o.id} className="">
                <Link
                  href={`/checkout/qr?order_id=${encodeURIComponent(o.id)}&from=orders`}
                  className="flex items-center justify-between gap-4"
                >
                  {/* Left: QR icon in bordered rounded box */}
                  <span className="border-border bg-background shrink-0 rounded-[10px] border p-2">
                    <QrCode className="h-8 w-8" />
                  </span>

                  {/* Middle: Name "Bestellung (date)" */}
                  <span className="min-w-0 flex-1 truncate text-base font-medium">Bestellung {dateLabel}</span>

                  {/* Right: Arrow in bordered rounded box */}
                  <span className="border-border bg-background shrink-0 rounded-[7px] border p-2">
                    <ArrowRight className="h-4 w-4" />
                  </span>
                </Link>
              </li>
            )
          })}
        </ul>
      )}

      <div className="fixed inset-x-0 bottom-0 z-50 p-4">
        <Button
          variant="outline"
          className="rounded-pill h-12 w-full bg-white text-base font-medium text-black"
          onClick={() => router.push("/")}
        >
          <ArrowLeft style={{ width: 20, height: 20 }} /> Zum Menü
        </Button>
      </div>
    </div>
  )
}
