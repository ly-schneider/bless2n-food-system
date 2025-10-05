"use client"
import { useEffect, useMemo, useState } from "react"
import { useAuthorizedFetch } from "@/hooks/use-authorized-fetch"
import {
  ChartContainer,
  ChartLegend,
  ChartLegendContent,
  ChartTooltip,
  ChartTooltipContent,
} from "@/components/ui/chart"
import {
  Bar,
  BarChart,
  CartesianGrid,
  Line,
  LineChart,
  XAxis,
  YAxis,
} from "recharts"

type OrderItem = { id: string; totalCents?: number | null; createdAt?: string }

export default function AdminDashboard() {
  const fetchAuth = useAuthorizedFetch()
  const [error, setError] = useState<string | null>(null)

  const [series14d, setSeries14d] = useState<
    { date: string; orders: number; revenue: number }[]
  >([])
  const [lowStock, setLowStock] = useState<
    { id: string; name: string; qty?: number | null }[]
  >([])

  useEffect(() => {
    let cancelled = false
    ;(async () => {
      try {
        const now = new Date()
        const start = new Date(now)
        start.setDate(now.getDate() - 13)
        start.setHours(0, 0, 0, 0)
        const end = new Date(now)
        end.setDate(now.getDate() + 1)
        end.setHours(0, 0, 0, 0)

        const [ordersRes, productsRes] = await Promise.all([
          fetchAuth(
            `${process.env.NEXT_PUBLIC_API_BASE_URL}/v1/admin/orders?date_from=${encodeURIComponent(
              start.toISOString()
            )}&date_to=${encodeURIComponent(end.toISOString())}&limit=10000`
          ),
          fetchAuth(`${process.env.NEXT_PUBLIC_API_BASE_URL}/v1/products?limit=200`),
        ])

        // Build 14d timeseries for orders and revenue
        let items: OrderItem[] = []
        if (ordersRes.ok) {
          const data = (await ordersRes.json()) as { items?: OrderItem[] }
          items = data.items || []
        }

        const dayKey = (d: Date) => d.toISOString().slice(0, 10) // YYYY-MM-DD
        const byDay = new Map<string, { orders: number; revenue: number }>()
        for (let i = 0; i < 14; i++) {
          const d = new Date(start)
          d.setDate(start.getDate() + i)
          byDay.set(dayKey(d), { orders: 0, revenue: 0 })
        }
        for (const it of items) {
          const created = it.createdAt ? new Date(it.createdAt) : null
          if (!created) continue
          const key = dayKey(created)
          if (!byDay.has(key)) continue
          const rec = byDay.get(key)!
          rec.orders += 1
          rec.revenue += (it.totalCents ?? 0) / 100
        }
        const fmtDay = (iso: string) => {
          const d = new Date(iso)
          return d.toLocaleDateString("de-CH", { day: "2-digit", month: "2-digit" })
        }
        const series = Array.from(byDay.entries())
          .sort(([a], [b]) => (a < b ? -1 : 1))
          .map(([k, v]) => ({ date: fmtDay(k), orders: v.orders, revenue: Number(v.revenue.toFixed(2)) }))

        // Low stock
        let low: { id: string; name: string; qty?: number | null }[] = []
        if (productsRes.ok) {
          const pr = (await productsRes.json()) as {
            items?: Array<{
              id: string
              name: string
              isLowStock?: boolean
              availableQuantity?: number | null
            }>
          }
          const items = pr.items || []
          low = items
            .filter((p) => p.isLowStock)
            .map((p) => ({ id: p.id, name: p.name, qty: p.availableQuantity ?? null }))
        }

        if (!cancelled) {
          setSeries14d(series)
          setLowStock(low)
        }
      } catch (e: unknown) {
        const msg = e instanceof Error ? e.message : "Dashboard laden fehlgeschlagen"
        if (!cancelled) setError(msg)
      }
    })()
    return () => {
      cancelled = true
    }
  }, [fetchAuth])

  const lowList = useMemo(() => lowStock.slice(0, 5), [lowStock])

  const chartConfig = {
    orders: { label: "Bestellungen", color: "#0ea5e9" }, // sky-500
    revenue: { label: "Umsatz (CHF)", color: "#a78bfa" }, // violet-400
    qty: { label: "Menge", color: "#f59e0b" }, // amber-500
  } as const

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-semibold">Adminbereich</h1>
      {error && <div className="text-red-600 text-sm">{error}</div>}

      <section className="grid grid-cols-1 gap-6 xl:grid-cols-2">
        <div className="rounded-lg border border-gray-200 p-4">
          <h2 className="mb-2 text-sm font-medium text-gray-600">Bestellungen (letzte 14 Tage)</h2>
          <ChartContainer config={chartConfig} className="h-60 w-full">
            <BarChart data={series14d} margin={{ left: 8, right: 8 }}>
              <CartesianGrid vertical={false} strokeDasharray="3 3" />
              <XAxis dataKey="date" tickLine={false} axisLine={false} />
              <YAxis allowDecimals={false} width={28} axisLine={false} tickLine={false} />
              <ChartTooltip cursor={{ opacity: 0.1 }} content={<ChartTooltipContent />} />
              <Bar dataKey="orders" fill="var(--color-orders)" radius={4} />
              <ChartLegend content={<ChartLegendContent />} />
            </BarChart>
          </ChartContainer>
        </div>

        <div className="rounded-lg border border-gray-200 p-4">
          <h2 className="mb-2 text-sm font-medium text-gray-600">Umsatz (letzte 14 Tage)</h2>
          <ChartContainer config={chartConfig} className="h-60 w-full">
            <LineChart data={series14d} margin={{ left: 8, right: 8 }}>
              <CartesianGrid vertical={false} strokeDasharray="3 3" />
              <XAxis dataKey="date" tickLine={false} axisLine={false} />
              <YAxis width={40} axisLine={false} tickLine={false} />
              <ChartTooltip
                cursor={{ strokeOpacity: 0.1 }}
                content={
                  <ChartTooltipContent
                    formatter={(value) => (
                      <span className="font-mono">{new Intl.NumberFormat("de-CH", { style: "currency", currency: "CHF" }).format(Number(value))}</span>
                    )}
                  />
                }
              />
              <Line
                type="monotone"
                dataKey="revenue"
                stroke="var(--color-revenue)"
                strokeWidth={2}
                dot={false}
                activeDot={{ r: 3 }}
              />
              <ChartLegend content={<ChartLegendContent />} />
            </LineChart>
          </ChartContainer>
        </div>
      </section>

      <section className="rounded-lg border border-gray-200 p-4">
        <h2 className="mb-2 text-sm font-medium text-gray-600">Artikel mit niedrigem Bestand</h2>
        {lowList.length === 0 ? (
          <div className="text-sm text-gray-500">Keine Artikel mit niedrigem Bestand.</div>
        ) : (
          <ChartContainer config={chartConfig} className="h-60 w-full">
            <BarChart
              data={lowList.map((l) => ({ name: l.name, qty: typeof l.qty === "number" ? l.qty : 0 }))}
              layout="vertical"
              margin={{ left: 8, right: 8 }}
            >
              <CartesianGrid horizontal={false} strokeDasharray="3 3" />
              <XAxis type="number" axisLine={false} tickLine={false} />
              <YAxis
                type="category"
                dataKey="name"
                width={160}
                axisLine={false}
                tickLine={false}
                tick={{ fontSize: 12 }}
              />
              <ChartTooltip cursor={{ opacity: 0.1 }} content={<ChartTooltipContent />} />
              <Bar dataKey="qty" fill="var(--color-qty)" radius={4} />
            </BarChart>
          </ChartContainer>
        )}
      </section>
    </div>
  )
}
