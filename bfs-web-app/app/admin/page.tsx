"use client"

import { ArrowRight, RadioTower } from "lucide-react"
import Image from "next/image"
import Link from "next/link"
import { useEffect, useMemo, useState } from "react"
import { Badge } from "@/components/ui/badge"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { useAuthorizedFetch } from "@/hooks/use-authorized-fetch"
import { cn, formatChf } from "@/lib/utils"
import type { AdminOpsOverview, DashboardStatusLevel } from "@/types"
import { OrdersByOriginChart } from "./_components/orders-by-origin-chart"
import { OrdersByStatusChart } from "./_components/orders-by-status-chart"
import { PaymentMethodsChart } from "./_components/payment-methods-chart"
import { RecentCancellations } from "./_components/recent-cancellations"
import { RevenueByHourChart } from "./_components/revenue-by-hour-chart"
import { StatCard } from "./_components/stat-card"
import { TopProductsChart } from "./_components/top-products-chart"

type OrderPayment = {
  method?: string
  amountCents?: number
}

type OrderLine = {
  id: string
  productId: string
  title: string
  quantity: number
  unitPriceCents: number
  lineType?: string
  parentLineId?: string | null
}

type OrderItem = {
  id: string
  status?: string
  origin?: string
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

type EventDay = {
  year: number
  month: number
  day: number
  orderCount: number
}

const LOW_STOCK_THRESHOLD = 10

const PAYMENT_METHOD_LABELS: Record<string, string> = {
  CASH: "Bargeld",
  CARD: "Karte",
  TWINT: "TWINT",
  GRATIS_GUEST: "Gratis (Gast)",
  GRATIS_VIP: "Gratis (VIP)",
  GRATIS_STAFF: "Gratis (Personal)",
  GRATIS_100CLUB: "Gratis (100 Club)",
}

const ORIGIN_LABELS: Record<string, string> = {
  shop: "Shop",
  pos: "POS",
}

function isGratisMethod(method?: string) {
  return method?.startsWith("GRATIS_") ?? false
}

function eventKey(event: EventDay) {
  return `${event.year}-${String(event.month).padStart(2, "0")}-${String(event.day).padStart(2, "0")}`
}

function formatEventLabel(event: EventDay | undefined) {
  if (!event) return "–"
  return new Date(event.year, event.month - 1, event.day).toLocaleDateString("de-CH", {
    weekday: "short",
    day: "2-digit",
    month: "2-digit",
    year: "numeric",
  })
}

function statusClasses(status: DashboardStatusLevel) {
  if (status === "red") return "border-red-300 bg-red-50 text-red-800"
  if (status === "yellow") return "border-amber-300 bg-amber-50 text-amber-900"
  return "border-emerald-300 bg-emerald-50 text-emerald-900"
}

function severityCopy(status: DashboardStatusLevel) {
  if (status === "red") return "Kritisch"
  if (status === "yellow") return "Beobachten"
  return "Stabil"
}

export default function AdminDashboard() {
  const fetchAuth = useAuthorizedFetch()
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState(true)

  const [orders, setOrders] = useState<OrderItem[]>([])
  const [prevDayOrders, setPrevDayOrders] = useState<OrderItem[]>([])
  const [products, setProducts] = useState<ProductWithStock[]>([])

  const [events, setEvents] = useState<EventDay[]>([])
  const [selectedEventKey, setSelectedEventKey] = useState("")
  const [eventsLoading, setEventsLoading] = useState(true)

  const [opsOverview, setOpsOverview] = useState<AdminOpsOverview | null>(null)
  const [opsLoading, setOpsLoading] = useState(true)
  const [opsError, setOpsError] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false
    setEventsLoading(true)
    fetchAuth("/api/v1/events")
      .then((res) => (res.ok ? res.json() : Promise.reject(new Error("Failed to load events"))))
      .then((data) => {
        if (!cancelled) {
          const typedData = data as { items?: EventDay[] }
          const nextEvents = typedData.items || []
          setEvents(nextEvents)
          setSelectedEventKey((current) => current || (nextEvents[0] ? eventKey(nextEvents[0]) : ""))
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
    let cancelled = false
    setOpsLoading(true)
    setOpsError(null)

    fetchAuth("/api/v1/dashboard/ops-overview")
      .then((res) => (res.ok ? res.json() : Promise.reject(new Error(`HTTP ${res.status}`))))
      .then((data) => {
        if (!cancelled) {
          setOpsOverview(data as AdminOpsOverview)
          setOpsLoading(false)
        }
      })
      .catch((err: unknown) => {
        if (!cancelled) {
          setOpsOverview(null)
          setOpsError(err instanceof Error ? err.message : "Live-Betrieb konnte nicht geladen werden.")
          setOpsLoading(false)
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

    const currentIndex = Math.max(
      0,
      events.findIndex((event) => eventKey(event) === selectedEventKey)
    )
    const event = events[currentIndex]
    if (!event) {
      setLoading(false)
      return
    }

    let cancelled = false
    setLoading(true)
    setError(null)

    ;(async () => {
      try {
        const from = new Date(event.year, event.month - 1, event.day)
        const to = new Date(event.year, event.month - 1, event.day + 1)
        const dateParams = `date_from=${encodeURIComponent(from.toISOString())}&date_to=${encodeURIComponent(to.toISOString())}`

        const prevEvent = events[currentIndex + 1]
        const prevDateParams = prevEvent
          ? `date_from=${encodeURIComponent(new Date(prevEvent.year, prevEvent.month - 1, prevEvent.day).toISOString())}&date_to=${encodeURIComponent(new Date(prevEvent.year, prevEvent.month - 1, prevEvent.day + 1).toISOString())}`
          : null

        const fetches: Promise<Response>[] = [fetchAuth(`/api/v1/orders?${dateParams}`), fetchAuth("/api/v1/products")]
        if (prevDateParams) {
          fetches.push(fetchAuth(`/api/v1/orders?status=paid&${prevDateParams}`))
        }

        const [ordersRes, productsRes, prevRes] = await Promise.all(fetches)

        let fetchedOrders: OrderItem[] = []
        if (ordersRes?.ok) {
          const data = (await ordersRes.json()) as { items?: OrderItem[] }
          fetchedOrders = data.items || []
        }

        let fetchedProducts: ProductWithStock[] = []
        if (productsRes?.ok) {
          const data = (await productsRes.json()) as { items?: ProductWithStock[] }
          fetchedProducts = data.items || []
        }

        let fetchedPrevOrders: OrderItem[] = []
        if (prevRes?.ok) {
          const data = (await prevRes.json()) as { items?: OrderItem[] }
          fetchedPrevOrders = data.items || []
        }

        if (!cancelled) {
          setOrders(fetchedOrders)
          setProducts(fetchedProducts)
          setPrevDayOrders(fetchedPrevOrders)
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
  }, [fetchAuth, events, selectedEventKey])

  const paidOrders = useMemo(() => orders.filter((o) => o.status === "paid"), [orders])

  const paidNonGratisOrders = useMemo(
    () => paidOrders.filter((o) => !isGratisMethod(o.payments?.[0]?.method)),
    [paidOrders]
  )

  const totalRevenue = useMemo(
    () => paidNonGratisOrders.reduce((sum, o) => sum + (o.totalCents ?? 0), 0),
    [paidNonGratisOrders]
  )

  const totalOrders = paidOrders.length

  const prevDayRevenue = useMemo(
    () =>
      prevDayOrders
        .filter((o) => !isGratisMethod(o.payments?.[0]?.method))
        .reduce((sum, o) => sum + (o.totalCents ?? 0), 0),
    [prevDayOrders]
  )

  const prevDayOrderCount = prevDayOrders.length

  const averageOrderValue = useMemo(
    () => (paidNonGratisOrders.length > 0 ? Math.round(totalRevenue / paidNonGratisOrders.length) : 0),
    [totalRevenue, paidNonGratisOrders.length]
  )

  const prevDayAvgOrderValue = useMemo(() => {
    const prevNonGratis = prevDayOrders.filter((o) => !isGratisMethod(o.payments?.[0]?.method))
    return prevNonGratis.length > 0
      ? Math.round(prevNonGratis.reduce((s, o) => s + (o.totalCents ?? 0), 0) / prevNonGratis.length)
      : 0
  }, [prevDayOrders])

  const cancellationCount = useMemo(
    () => orders.filter((o) => o.status === "cancelled" || o.status === "refunded").length,
    [orders]
  )

  const revenueByHour = useMemo(() => {
    const byHour: Record<number, number> = {}
    for (const o of paidNonGratisOrders) {
      if (!o.createdAt) continue
      const h = new Date(o.createdAt).getHours()
      byHour[h] = (byHour[h] || 0) + (o.totalCents ?? 0)
    }
    if (Object.keys(byHour).length === 0) return []
    const minH = Math.min(...Object.keys(byHour).map(Number))
    const maxH = Math.max(...Object.keys(byHour).map(Number))
    const result = []
    for (let h = minH; h <= maxH; h++) {
      result.push({ hour: `${h.toString().padStart(2, "0")}:00`, revenueCents: byHour[h] || 0 })
    }
    return result
  }, [paidNonGratisOrders])

  const paymentMethodCounts = useMemo(() => {
    const counts: Record<string, number> = {}
    for (const o of paidOrders) {
      const method = o.payments?.[0]?.method ?? "UNKNOWN"
      counts[method] = (counts[method] || 0) + 1
    }
    return Object.entries(counts)
      .map(([method, count]) => ({
        method,
        label: PAYMENT_METHOD_LABELS[method] ?? method,
        count,
      }))
      .sort((a, b) => b.count - a.count)
  }, [paidOrders])

  const ordersByOrigin = useMemo(() => {
    const counts: Record<string, number> = {}
    for (const o of paidOrders) {
      const origin = o.origin || "shop"
      counts[origin] = (counts[origin] || 0) + 1
    }
    return Object.entries(counts)
      .map(([origin, count]) => ({
        origin,
        label: ORIGIN_LABELS[origin] ?? origin,
        count,
      }))
      .sort((a, b) => b.count - a.count)
  }, [paidOrders])

  const topProductsByUnits = useMemo(() => {
    const agg: Record<string, { title: string; value: number }> = {}
    for (const o of paidOrders) {
      for (const line of o.lines ?? []) {
        if (line.lineType === "component" || line.parentLineId) continue
        const entry = agg[line.productId] ?? (agg[line.productId] = { title: line.title, value: 0 })
        entry.value += line.quantity
      }
    }
    return Object.values(agg)
      .sort((a, b) => b.value - a.value)
      .slice(0, 15)
  }, [paidOrders])

  const topProductsByRevenue = useMemo(() => {
    const agg: Record<string, { title: string; value: number }> = {}
    for (const o of paidOrders) {
      for (const line of o.lines ?? []) {
        if (line.lineType === "component" || line.parentLineId) continue
        const entry = agg[line.productId] ?? (agg[line.productId] = { title: line.title, value: 0 })
        entry.value += line.quantity * (line.unitPriceCents || 0)
      }
    }
    return Object.values(agg)
      .sort((a, b) => b.value - a.value)
      .slice(0, 15)
  }, [paidOrders])

  const ordersByStatusByHour = useMemo(() => {
    const byHour: Record<number, { paid: number; pending: number; cancelled: number; refunded: number }> = {}
    for (const o of orders) {
      if (!o.createdAt) continue
      const h = new Date(o.createdAt).getHours()
      if (!byHour[h]) byHour[h] = { paid: 0, pending: 0, cancelled: 0, refunded: 0 }
      const status = o.status as keyof (typeof byHour)[number]
      if (status && status in byHour[h]) byHour[h][status]++
    }
    if (Object.keys(byHour).length === 0) return []
    const minH = Math.min(...Object.keys(byHour).map(Number))
    const maxH = Math.max(...Object.keys(byHour).map(Number))
    const result = []
    for (let h = minH; h <= maxH; h++) {
      result.push({
        hour: `${h.toString().padStart(2, "0")}:00`,
        ...(byHour[h] || { paid: 0, pending: 0, cancelled: 0, refunded: 0 }),
      })
    }
    return result
  }, [orders])

  const recentCancellations = useMemo(
    () =>
      orders
        .filter((o) => o.status === "cancelled" || o.status === "refunded")
        .sort((a, b) => new Date(b.createdAt ?? 0).getTime() - new Date(a.createdAt ?? 0).getTime())
        .slice(0, 10)
        .map((o) => ({
          id: o.id,
          status: o.status!,
          totalCents: o.totalCents ?? 0,
          origin: o.origin || "shop",
          createdAt: o.createdAt!,
        })),
    [orders]
  )

  const lowStockProducts = useMemo(() => {
    return products
      .filter((p) => p.type !== "menu" && typeof p.stock === "number" && p.stock <= LOW_STOCK_THRESHOLD)
      .sort((a, b) => (a.stock ?? 0) - (b.stock ?? 0))
      .slice(0, 10)
  }, [products])

  function comparisonLabel(current: number, previous: number): { label: string; positive: boolean } | null {
    if (previous === 0) return null
    const delta = current - previous
    const pct = ((delta / previous) * 100).toFixed(0)
    const sign = delta >= 0 ? "+" : ""
    return { label: `${sign}${pct}% zum Vortag`, positive: delta >= 0 }
  }

  const hasPrevDay = prevDayOrders.length > 0
  const selectedEvent = events.find((event) => eventKey(event) === selectedEventKey) || events[0]

  return (
    <div className="space-y-6">
      <div className="space-y-1">
        <h1 className="text-2xl font-semibold">Adminbereich</h1>
        <p className="text-muted-foreground text-sm">Live-Betrieb und Event-Auswertung in einer Ansicht.</p>
      </div>

      <Card className="overflow-hidden border-slate-200/80">
        <CardHeader className="pb-4">
          <div className="flex flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
            <div className="space-y-1">
              <div className="flex items-center gap-2">
                <Badge
                  variant="outline"
                  className={cn("rounded-full border px-3 py-1", statusClasses(opsOverview?.overallStatus || "green"))}
                >
                  <RadioTower className="mr-1 size-3.5" />
                  {severityCopy(opsOverview?.overallStatus || "green")}
                </Badge>
                <span className="text-muted-foreground text-sm">Live-Betrieb jetzt</span>
              </div>
              <CardTitle className="text-xl">Operative Übersicht</CardTitle>
              <CardDescription>Systemstatus und Stationsdruck separat von der Event-Auswertung.</CardDescription>
            </div>
            <Link href="/admin/stations" className="text-sm font-medium underline underline-offset-4">
              Stationen verwalten
            </Link>
          </div>
        </CardHeader>
        <CardContent className="space-y-4">
          {opsError && <div className="text-sm text-red-600">{opsError}</div>}

          <div className="grid gap-3 md:grid-cols-2 xl:grid-cols-4">
            {(opsOverview?.statusChips || []).map((chip) => (
              <div key={chip.key} className={cn("rounded-2xl border px-4 py-3", statusClasses(chip.status))}>
                <div className="text-xs font-medium tracking-[0.16em] uppercase opacity-80">{chip.label}</div>
                <div className="mt-1 text-sm font-semibold">{chip.summary}</div>
              </div>
            ))}
            {opsLoading &&
              Array.from({ length: 4 }).map((_, index) => (
                <div key={index} className="rounded-2xl border border-dashed px-4 py-6 text-sm text-slate-500">
                  Lade Live-Status…
                </div>
              ))}
          </div>

          <div className="grid gap-3 md:grid-cols-2 xl:grid-cols-4">
            {!opsLoading && (opsOverview?.stations.length || 0) === 0 && (
              <div className="text-muted-foreground rounded-2xl border border-dashed px-4 py-8 text-sm md:col-span-2 xl:col-span-4">
                Keine freigegebenen Stationen vorhanden.
              </div>
            )}

            {(opsOverview?.stations || []).map((station) => (
              <Link key={station.id} href={`/admin/stations/${station.id}`}>
                <div className="rounded-2xl border border-slate-200 p-4 transition-shadow hover:shadow-sm">
                  <div className="flex items-start justify-between gap-3">
                    <div>
                      <div className="font-medium">{station.name}</div>
                      <div className="text-muted-foreground mt-1 text-sm">
                        {station.recentProductTitle
                          ? `${station.recentProductTitle} · ${station.recentProductQuantity}× zuletzt`
                          : "Keine aktuelle Produktbewegung"}
                      </div>
                    </div>
                    <Badge variant="outline" className={cn("rounded-full", statusClasses(station.status))}>
                      {severityCopy(station.status)}
                    </Badge>
                  </div>

                  <div className="mt-4 grid grid-cols-3 gap-2 text-sm">
                    <div className="rounded-xl bg-slate-100 px-3 py-2">
                      <div className="text-muted-foreground text-xs">Offene</div>
                      <div className="mt-1 font-semibold">{station.openOrders}</div>
                    </div>
                    <div className="rounded-xl bg-slate-100 px-3 py-2">
                      <div className="text-muted-foreground text-xs">Backlog</div>
                      <div className="mt-1 font-semibold">{station.backlog}</div>
                    </div>
                    <div className="rounded-xl bg-slate-100 px-3 py-2">
                      <div className="text-muted-foreground text-xs">Median</div>
                      <div className="mt-1 font-semibold">{station.medianThroughputMinutes}m</div>
                    </div>
                  </div>

                  <div className="text-muted-foreground mt-4 flex items-center justify-between text-sm">
                    <span>Details öffnen</span>
                    <ArrowRight className="size-4" />
                  </div>
                </div>
              </Link>
            ))}
          </div>
        </CardContent>
      </Card>

      {error && <div className="text-sm text-red-600">{error}</div>}

      <section className="space-y-3">
        <div className="flex flex-col gap-3 lg:flex-row lg:items-end lg:justify-between">
          <div>
            <h2 className="text-xl font-semibold">Event-Auswertung</h2>
            <p className="text-muted-foreground mt-1 text-sm">Historische Kennzahlen und Vergleiche pro Eventtag.</p>
          </div>
          <div className="w-full sm:w-[280px]">
            <Select
              value={selectedEventKey || "__empty"}
              onValueChange={(next) => setSelectedEventKey(next === "__empty" ? "" : next)}
              disabled={eventsLoading || events.length === 0}
            >
              <SelectTrigger>
                <SelectValue placeholder="Eventtag wählen" />
              </SelectTrigger>
              <SelectContent>
                {events.length === 0 ? (
                  <SelectItem value="__empty">Kein Event vorhanden</SelectItem>
                ) : (
                  events.map((event) => (
                    <SelectItem key={eventKey(event)} value={eventKey(event)}>
                      {formatEventLabel(event)} · {event.orderCount} Orders
                    </SelectItem>
                  ))
                )}
              </SelectContent>
            </Select>
          </div>
        </div>
        {!eventsLoading && selectedEvent && (
          <div className="text-muted-foreground text-sm">Ausgewählt: {formatEventLabel(selectedEvent)}</div>
        )}
      </section>

      <section className="grid grid-cols-2 gap-4 md:grid-cols-4">
        <StatCard
          title="Umsatz"
          value={formatChf(totalRevenue)}
          loading={loading}
          comparison={hasPrevDay ? comparisonLabel(totalRevenue, prevDayRevenue) : null}
        />
        <StatCard
          title="Bestellungen"
          value={String(totalOrders)}
          loading={loading}
          comparison={hasPrevDay ? comparisonLabel(totalOrders, prevDayOrderCount) : null}
        />
        <StatCard
          title="Ø Bestellwert"
          value={formatChf(averageOrderValue)}
          loading={loading}
          comparison={hasPrevDay ? comparisonLabel(averageOrderValue, prevDayAvgOrderValue) : null}
        />
        <StatCard title="Stornierungen" value={String(cancellationCount)} loading={loading} />
      </section>

      <section className="grid grid-cols-1 gap-4 lg:grid-cols-[2fr_1fr_1fr]">
        <RevenueByHourChart data={revenueByHour} loading={loading} />
        <PaymentMethodsChart data={paymentMethodCounts} loading={loading} />
        <OrdersByOriginChart data={ordersByOrigin} loading={loading} />
      </section>

      <section className="grid grid-cols-1 gap-4 md:grid-cols-2">
        <TopProductsChart title="Top Produkte nach Menge" data={topProductsByUnits} loading={loading} />
        <TopProductsChart
          title="Top Produkte nach Umsatz"
          data={topProductsByRevenue}
          loading={loading}
          formatAsCurrency
        />
      </section>

      <section className="grid grid-cols-1 gap-4 md:grid-cols-2">
        <OrdersByStatusChart data={ordersByStatusByHour} loading={loading} />
        <RecentCancellations data={recentCancellations} loading={loading} />
      </section>

      <section>
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
