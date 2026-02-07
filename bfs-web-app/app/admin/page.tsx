"use client"
import { ChevronLeft, ChevronRight } from "lucide-react"
import Image from "next/image"
import { useEffect, useMemo, useState } from "react"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog"
import { useAuthorizedFetch } from "@/hooks/use-authorized-fetch"
import { formatChf } from "@/lib/utils"

type OrderPayment = {
  method?: string
  amountCents?: number
}

type OrderLine = {
  id: string
  productId: string
  title: string
  quantity: number
  lineType?: string
  parentLineId?: string | null
}

type OrderItem = {
  id: string
  totalCents?: number | null
  createdAt?: string
  payments?: OrderPayment[] | null
  lines?: OrderLine[] | null
}

type ProductWithStock = {
  id: string
  name: string
  type?: string
  image?: string | null
  stock?: number | null
}

type EventMonth = {
  year: number
  month: number
  orderCount: number
}

const LOW_STOCK_THRESHOLD = 10

export default function AdminDashboard() {
  const fetchAuth = useAuthorizedFetch()
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState(true)

  const [orders, setOrders] = useState<OrderItem[]>([])
  const [products, setProducts] = useState<ProductWithStock[]>([])

  const [events, setEvents] = useState<EventMonth[]>([])
  const [currentEventIndex, setCurrentEventIndex] = useState(0)
  const [eventsLoading, setEventsLoading] = useState(true)

  useEffect(() => {
    let cancelled = false
    setEventsLoading(true)
    fetchAuth("/api/v1/events")
      .then((res) => (res.ok ? res.json() : Promise.reject(new Error("Failed to load events"))))
      .then((data) => {
        if (!cancelled) {
          const typedData = data as { items?: EventMonth[] }
          setEvents(typedData.items || [])
          setEventsLoading(false)
        }
      })
      .catch(() => {
        if (!cancelled) {
          setEventsLoading(false)
        }
      })
    return () => {
      cancelled = true
    }
  }, [fetchAuth])

  useEffect(() => {
    if (events.length === 0) {
      setLoading(false)
      return
    }
    const event = events[currentEventIndex]
    if (!event) {
      setLoading(false)
      return
    }
    let cancelled = false
    setLoading(true)
    ;(async () => {
      try {
        const from = new Date(event.year, event.month - 1, 1)
        const to = new Date(event.year, event.month, 1)

        const [ordersRes, productsRes] = await Promise.all([
          fetchAuth(
            `/api/v1/orders?date_from=${encodeURIComponent(from.toISOString())}&date_to=${encodeURIComponent(
              to.toISOString()
            )}`
          ),
          fetchAuth(`/api/v1/products`),
        ])

        let fetchedOrders: OrderItem[] = []
        if (ordersRes.ok) {
          const data = (await ordersRes.json()) as { items?: OrderItem[] }
          fetchedOrders = data.items || []
        }

        let fetchedProducts: ProductWithStock[] = []
        if (productsRes.ok) {
          const data = (await productsRes.json()) as { items?: ProductWithStock[] }
          fetchedProducts = data.items || []
        }

        if (!cancelled) {
          setOrders(fetchedOrders)
          setProducts(fetchedProducts)
          setLoading(false)
        }
      } catch (e: unknown) {
        const msg = e instanceof Error ? e.message : "Dashboard laden fehlgeschlagen"
        if (!cancelled) {
          setError(msg)
          setLoading(false)
        }
      }
    })()
    return () => {
      cancelled = true
    }
  }, [fetchAuth, events, currentEventIndex])

  const totalRevenue = useMemo(() => {
    return orders.reduce((sum, o) => sum + (o.totalCents ?? 0), 0)
  }, [orders])

  const totalOrders = orders.length

  const revenueByPaymentType = useMemo(() => {
    const byType: Record<string, number> = {}
    for (const order of orders) {
      const payment = order.payments?.[0]
      const method = payment?.method ?? "unbekannt"
      byType[method] = (byType[method] || 0) + (order.totalCents ?? 0)
    }
    return Object.entries(byType)
      .map(([method, cents]) => ({ method, cents }))
      .sort((a, b) => b.cents - a.cents)
  }, [orders])

  const productOrderCounts = useMemo(() => {
    const counts: Record<string, { title: string; count: number }> = {}
    for (const order of orders) {
      for (const line of order.lines ?? []) {
        if (line.lineType === "component" || line.parentLineId) continue
        const key = line.productId
        if (!counts[key]) {
          counts[key] = { title: line.title, count: 0 }
        }
        counts[key].count += line.quantity
      }
    }
    return Object.values(counts).sort((a, b) => b.count - a.count)
  }, [orders])

  const lowStockProducts = useMemo(() => {
    return products
      .filter((p) => p.type !== "menu" && typeof p.stock === "number" && p.stock <= LOW_STOCK_THRESHOLD)
      .sort((a, b) => (a.stock ?? 0) - (b.stock ?? 0))
      .slice(0, 10)
  }, [products])

  const paymentMethodLabels: Record<string, string> = {
    cash: "Bargeld",
    card: "Karte",
    twint: "TWINT",
    unbekannt: "Unbekannt",
  }

  const currentMonthLabel = useMemo(() => {
    if (events.length === 0) return "–"
    const event = events[currentEventIndex]
    if (!event) return "–"
    return new Date(event.year, event.month - 1).toLocaleDateString("de-CH", {
      month: "long",
      year: "numeric",
    })
  }, [events, currentEventIndex])

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-semibold">Adminbereich</h1>
      {error && <div className="text-sm text-red-600">{error}</div>}

      <div className="flex items-center gap-3">
        <span className="text-muted-foreground text-sm">Zeitraum:</span>
        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            size="icon"
            onClick={() => setCurrentEventIndex((i) => i + 1)}
            disabled={eventsLoading || currentEventIndex >= events.length - 1}
          >
            <ChevronLeft className="size-4" />
          </Button>
          <span className="min-w-[160px] text-center font-medium">{currentMonthLabel}</span>
          <Button
            variant="outline"
            size="icon"
            onClick={() => setCurrentEventIndex((i) => i - 1)}
            disabled={eventsLoading || currentEventIndex === 0}
          >
            <ChevronRight className="size-4" />
          </Button>
        </div>
      </div>

      <section className="grid grid-cols-1 gap-4 md:grid-cols-2">
        <Dialog>
          <div className="rounded-lg border border-gray-200">
            <div className="text-muted-foreground text-sm font-medium">Umsatz</div>
            <div className="mt-1 text-3xl font-semibold">
              {loading ? "–" : formatChf(totalRevenue)}
            </div>
            <DialogTrigger asChild>
              <Button variant="outline" size="sm" className="mt-3 gap-1 px-0" disabled={loading}>
                Details <ChevronRight className="size-4" />
              </Button>
            </DialogTrigger>
          </div>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Umsatz nach Zahlungsart</DialogTitle>
            </DialogHeader>
            <div className="space-y-2">
              {revenueByPaymentType.length === 0 ? (
                <p className="text-muted-foreground text-sm">Keine Daten vorhanden.</p>
              ) : (
                revenueByPaymentType.map(({ method, cents }) => (
                  <div key={method} className="flex items-center justify-between">
                    <span className="capitalize">{paymentMethodLabels[method] ?? method}</span>
                    <span className="font-medium">{formatChf(cents)}</span>
                  </div>
                ))
              )}
              <div className="mt-4 flex items-center justify-between border-t pt-3">
                <span className="font-medium">Gesamt</span>
                <span className="font-semibold">{formatChf(totalRevenue)}</span>
              </div>
            </div>
          </DialogContent>
        </Dialog>

        <Dialog>
          <div className="rounded-lg border border-gray-200">
            <div className="text-muted-foreground text-sm font-medium">Bestellungen</div>
            <div className="mt-1 text-3xl font-semibold">{loading ? "–" : totalOrders}</div>
            <DialogTrigger asChild>
              <Button variant="outline" size="sm" className="mt-3 gap-1 px-0" disabled={loading}>
                Details <ChevronRight className="size-4" />
              </Button>
            </DialogTrigger>
          </div>
          <DialogContent className="max-h-[80vh] overflow-y-auto">
            <DialogHeader>
              <DialogTitle>Produkte nach Bestellmenge</DialogTitle>
            </DialogHeader>
            <div className="space-y-2">
              {productOrderCounts.length === 0 ? (
                <p className="text-muted-foreground text-sm">Keine Daten vorhanden.</p>
              ) : (
                productOrderCounts.map(({ title, count }, idx) => (
                  <div key={idx} className="flex items-center justify-between">
                    <span className="truncate pr-4">{title}</span>
                    <span className="shrink-0 font-medium">{count}×</span>
                  </div>
                ))
              )}
            </div>
          </DialogContent>
        </Dialog>
      </section>

      <section className="mt-12">
        <h2 className="text-xl font-semibold">Artikel mit niedrigem Bestand</h2>
        <p className="text-muted-foreground mt-1 text-sm">Produkte mit weniger als {LOW_STOCK_THRESHOLD} Einheiten</p>
        {loading ? (
          <div className="text-muted-foreground mt-4 text-sm">Lade…</div>
        ) : lowStockProducts.length === 0 ? (
          <div className="mt-4 text-sm text-green-700">Alle Produkte haben ausreichend Bestand.</div>
        ) : (
          <div className="mt-4 flex flex-col gap-3">
            {lowStockProducts.map((p) => (
              <div key={p.id} className="rounded-xl border p-3">
                <div className="flex items-center gap-3">
                  <div className="relative h-10 w-10 shrink-0 overflow-hidden rounded-lg bg-[#cec9c6]">
                    {p.image && (
                      <Image
                        src={p.image}
                        alt={"Produktbild von " + p.name}
                        fill
                        sizes="40px"
                        quality={90}
                        unoptimized={p.image.includes("localhost:10000")}
                        className="h-full w-full rounded-lg object-cover"
                      />
                    )}
                  </div>
                  <div className="min-w-0 flex-1">
                    <div className="flex items-center justify-between gap-3">
                      <p className="min-w-0 truncate font-medium">{p.name}</p>
                      <div className="shrink-0 text-right">
                        <p className="text-muted-foreground text-sm">Bestand</p>
                        <p
                          className={`font-medium ${(p.stock ?? 0) <= 0 ? "text-red-600" : (p.stock ?? 0) <= 5 ? "text-orange-600" : ""}`}
                        >
                          {p.stock ?? 0}
                        </p>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}
      </section>
    </div>
  )
}
