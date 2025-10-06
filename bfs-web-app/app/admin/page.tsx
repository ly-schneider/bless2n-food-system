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
import { Calendar } from "@/components/ui/calendar"
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover"
import { Button } from "@/components/ui/button"
import { Calendar as CalendarIcon } from "lucide-react"
import type { DateRange } from "react-day-picker"
import { de } from "date-fns/locale"

type OrderItem = { id: string; totalCents?: number | null; createdAt?: string }

export default function AdminDashboard() {
  const fetchAuth = useAuthorizedFetch()
  const [error, setError] = useState<string | null>(null)

  const [series, setSeries] = useState<
    { date: string; orders: number; revenue: number }[]
  >([])
  const [lowStock, setLowStock] = useState<
    { id: string; name: string; qty?: number | null }[]
  >([])

  // Date range: default to last 30 days (inclusive today)
  const [range, setRange] = useState<DateRange>(() => {
    const now = new Date()
    const from = new Date(now)
    from.setDate(now.getDate() - 29)
    from.setHours(0, 0, 0, 0)
    const to = new Date(now)
    to.setHours(0, 0, 0, 0)
    return { from, to }
  })

  useEffect(() => {
    let cancelled = false
    ;(async () => {
      try {
        // Resolve date range and compute [start, endExclusive)
        const now = new Date()
        const today = new Date(now)
        today.setHours(0, 0, 0, 0)
        const start = new Date(range?.from ?? now)
        start.setHours(0, 0, 0, 0)
        const inclusiveTo = new Date(range?.to ?? today)
        inclusiveTo.setHours(0, 0, 0, 0)
        // Clamp to not go into the future
        if (inclusiveTo.getTime() > today.getTime()) {
          inclusiveTo.setTime(today.getTime())
        }
        const end = new Date(inclusiveTo)
        end.setDate(inclusiveTo.getDate() + 1)
        end.setHours(0, 0, 0, 0)

        const [ordersRes, productsRes] = await Promise.all([
          fetchAuth(
            `${process.env.NEXT_PUBLIC_API_BASE_URL}/v1/admin/orders?date_from=${encodeURIComponent(
              start.toISOString()
            )}&date_to=${encodeURIComponent(end.toISOString())}&limit=10000`
          ),
          fetchAuth(`${process.env.NEXT_PUBLIC_API_BASE_URL}/v1/products?limit=200`),
        ])

        // Build timeseries for orders and revenue
        let items: OrderItem[] = []
        if (ordersRes.ok) {
          const data = (await ordersRes.json()) as { items?: OrderItem[] }
          items = data.items || []
        }

        const dayKey = (d: Date) => d.toISOString().slice(0, 10) // YYYY-MM-DD
        const byDay = new Map<string, { orders: number; revenue: number }>()
        // number of days between start (inclusive) and end (exclusive)
        const millisPerDay = 24 * 60 * 60 * 1000
        const totalDays = Math.max(0, Math.round((end.getTime() - start.getTime()) / millisPerDay))
        for (let i = 0; i < totalDays; i++) {
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
        const builtSeries = Array.from(byDay.entries())
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
          setSeries(builtSeries)
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
  }, [fetchAuth, range.from?.getTime(), range.to?.getTime()])

  const lowList = useMemo(() => lowStock.slice(0, 5), [lowStock])

  const chartConfig = {
    orders: { label: "Bestellungen", color: "#0ea5e9" }, // sky-500
    revenue: { label: "Umsatz (CHF)", color: "#a78bfa" }, // violet-400
    qty: { label: "Menge", color: "#f59e0b" }, // amber-500
  } as const

  const fmtRangeLabel = useMemo(() => {
    const from = range?.from
    const to = range?.to
    if (!from && !to) return "Zeitraum wählen"
    const fmt = (d: Date) => d.toLocaleDateString("de-CH")
    if (from && !to) return `${fmt(from)} –`
    if (!from && to) return `– ${fmt(to)}`
    return `${fmt(from!)} – ${fmt(to!)}`
  }, [range])

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-semibold">Adminbereich</h1>
      {error && <div className="text-red-600 text-sm">{error}</div>}

      <div className="flex items-center gap-3">
        <span className="text-sm text-muted-foreground">Zeitraum:</span>
        <Popover>
          <PopoverTrigger asChild>
            <Button variant="outline" className="justify-start gap-2">
              <CalendarIcon className="size-4" />
              <span className="font-normal">{fmtRangeLabel}</span>
            </Button>
          </PopoverTrigger>
          <PopoverContent className="w-auto p-0" align="start">
            <Calendar
              mode="range"
              numberOfMonths={2}
              selected={range}
              onSelect={(r) => r && setRange(r)}
              defaultMonth={range?.from ?? new Date()}
              captionLayout="dropdown"
              locale={de}
              formatters={{
                formatMonthDropdown: (date) => date.toLocaleString("de-CH", { month: "short" }),
              }}
              disabled={{ after: (() => { const d = new Date(); d.setHours(0,0,0,0); return d })() }}
            />
          </PopoverContent>
        </Popover>
      </div>

      <section className="grid grid-cols-1 gap-6 xl:grid-cols-2">
        <div className="rounded-lg border border-gray-200 p-4">
          <h2 className="mb-2 text-sm font-medium text-gray-600">Bestellungen</h2>
          <ChartContainer config={chartConfig} className="h-60 w-full">
            <BarChart data={series} margin={{ left: 8, right: 8 }}>
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
          <h2 className="mb-2 text-sm font-medium text-gray-600">Umsatz</h2>
          <ChartContainer config={chartConfig} className="h-60 w-full">
            <LineChart data={series} margin={{ left: 8, right: 8 }}>
              <CartesianGrid vertical={false} strokeDasharray="3 3" />
              <XAxis dataKey="date" tickLine={false} axisLine={false} />
              <YAxis width={40} axisLine={false} tickLine={false} />
              <ChartTooltip
                cursor={{ strokeOpacity: 0.1 }}
                content={
                  <ChartTooltipContent
                    formatter={(value) => (
                      <span>{new Intl.NumberFormat("de-CH", { style: "currency", currency: "CHF" }).format(Number(value))}</span>
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
