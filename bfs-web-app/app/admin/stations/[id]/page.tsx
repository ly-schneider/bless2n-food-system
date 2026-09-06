"use client"

import { Clock3 } from "lucide-react"
import Link from "next/link"
import { useParams } from "next/navigation"
import { useEffect, useMemo, useState } from "react"
import { Bar, BarChart, CartesianGrid, XAxis, YAxis } from "recharts"
import { TopProductsChart } from "@/app/admin/_components/top-products-chart"
import { Badge } from "@/components/ui/badge"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { type ChartConfig, ChartContainer, ChartTooltip } from "@/components/ui/chart"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { useAuthorizedFetch } from "@/hooks/use-authorized-fetch"
import { cn } from "@/lib/utils"
import type { DashboardStatusLevel, StationDetailSummary } from "@/types"

function statusClasses(status: DashboardStatusLevel) {
  if (status === "red") return "border-red-300 bg-red-50 text-red-800"
  if (status === "yellow") return "border-amber-300 bg-amber-50 text-amber-900"
  return "border-emerald-300 bg-emerald-50 text-emerald-900"
}

const throughputChartConfig = {
  value: { label: "Fertiggestellt", color: "var(--chart-2)" },
} satisfies ChartConfig

export default function AdminStationDetailPage() {
  const fetchAuth = useAuthorizedFetch()
  const params = useParams<{ id: string }>()
  const stationId = params.id || ""
  const [summary, setSummary] = useState<StationDetailSummary | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (!stationId) return
    let cancelled = false

    async function load() {
      setLoading(true)
      setError(null)
      try {
        const res = await fetchAuth(`/api/v1/stations/${stationId}/summary`)
        if (!res.ok) throw new Error(`HTTP ${res.status}`)
        const data = (await res.json()) as StationDetailSummary
        if (!cancelled) {
          setSummary(data)
        }
      } catch (err: unknown) {
        if (!cancelled) {
          setSummary(null)
          setError(err instanceof Error ? err.message : "Stationsansicht konnte nicht geladen werden.")
        }
      } finally {
        if (!cancelled) {
          setLoading(false)
        }
      }
    }

    void load()
    return () => {
      cancelled = true
    }
  }, [fetchAuth, stationId])

  const topProductsData = useMemo(
    () => (summary?.topProducts || []).map((item) => ({ title: item.title, value: item.quantity })),
    [summary]
  )
  const throughputData = useMemo(
    () => (summary?.throughputByHour || []).map((item) => ({ hour: item.label, value: item.value })),
    [summary]
  )

  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-3">
        <Link href="/admin" className="text-muted-foreground text-sm underline underline-offset-4">
          Zurück zum Adminbereich
        </Link>
        <div className="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
          <div className="space-y-1">
            <div className="flex items-center gap-2">
              <h1 className="text-2xl font-semibold">{summary?.station.name || "Station"}</h1>
              <Badge variant="outline" className={cn("rounded-full", statusClasses(summary?.station.status || "green"))}>
                {summary?.station.status || "green"}
              </Badge>
            </div>
            <p className="text-muted-foreground text-sm">
              Aktuelle Queue, Durchsatz seit Tagesbeginn und relevante Produkte der letzten Stunde.
            </p>
          </div>
        </div>
      </div>

      {error && (
        <div className="rounded-2xl border border-red-300 bg-red-50 px-4 py-3 text-sm text-red-700">{error}</div>
      )}

      <div className="grid gap-4 md:grid-cols-3">
        <Card>
          <CardHeader className="pb-2">
            <CardDescription>Offene Orders</CardDescription>
            <CardTitle className="text-3xl">{loading ? "–" : summary?.station.openOrders || 0}</CardTitle>
          </CardHeader>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardDescription>Backlog</CardDescription>
            <CardTitle className="text-3xl">{loading ? "–" : summary?.station.backlog || 0}</CardTitle>
          </CardHeader>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardDescription>Median-Durchlaufzeit</CardDescription>
            <CardTitle className="text-3xl">
              {loading ? "–" : `${summary?.station.medianThroughputMinutes || 0}m`}
            </CardTitle>
          </CardHeader>
        </Card>
      </div>

      <div className="grid gap-4 xl:grid-cols-[1.1fr_0.9fr]">
        <Card className="rounded-[24px]">
          <CardHeader>
            <CardTitle className="flex items-center gap-2 text-lg">
              <Clock3 className="size-4" />
              Aktuelle Queue
            </CardTitle>
            <CardDescription>Die ältesten und dichtesten offenen Orders zuerst.</CardDescription>
          </CardHeader>
          <CardContent>
            {(summary?.queue.length || 0) === 0 && !loading ? (
              <div className="text-muted-foreground rounded-2xl border border-dashed px-4 py-10 text-center text-sm">
                Keine offene Queue seit Tagesbeginn.
              </div>
            ) : (
              <div className="overflow-x-auto">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Order</TableHead>
                      <TableHead>Alter</TableHead>
                      <TableHead>Items</TableHead>
                      <TableHead>Menge</TableHead>
                      <TableHead>Artikel</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {(summary?.queue || []).map((item) => (
                      <TableRow key={item.orderId}>
                        <TableCell>
                          <Link href={`/admin/orders/${item.orderId}`} className="underline underline-offset-4">
                            {item.orderId.slice(0, 8)}…
                          </Link>
                        </TableCell>
                        <TableCell>{item.ageMinutes}m</TableCell>
                        <TableCell>{item.pendingItems}</TableCell>
                        <TableCell>{item.pendingQuantity}</TableCell>
                        <TableCell className="max-w-[320px] truncate">{item.titles.join(", ")}</TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </div>
            )}
          </CardContent>
        </Card>

        <div className="space-y-4">
          <Card className="rounded-[24px]">
            <CardHeader>
              <CardTitle className="flex items-center gap-2 text-lg">
                <Clock3 className="size-4" />
                Fertigstellungen pro Stunde
              </CardTitle>
              <CardDescription>Abgeschlossene Mengen dieser Station seit Tagesbeginn.</CardDescription>
            </CardHeader>
            <CardContent>
              {loading ? (
                <div className="text-muted-foreground flex h-[220px] items-center justify-center text-sm">Lade…</div>
              ) : throughputData.length === 0 ? (
                <div className="text-muted-foreground flex h-[220px] items-center justify-center rounded-2xl border border-dashed text-sm">
                  Noch keine Fertigstellungen vorhanden.
                </div>
              ) : (
                <ChartContainer config={throughputChartConfig} className="h-[220px] w-full">
                  <BarChart data={throughputData} margin={{ top: 4, right: 4, bottom: 0, left: 0 }}>
                    <CartesianGrid vertical={false} strokeDasharray="3 3" />
                    <XAxis dataKey="hour" tickLine={false} axisLine={false} fontSize={12} />
                    <YAxis tickLine={false} axisLine={false} fontSize={12} width={32} />
                    <ChartTooltip
                      content={({ active, payload }) => {
                        if (!active || !payload?.length || !payload[0]) return null
                        const point = payload[0].payload as { hour: string; value: number }
                        return (
                          <div className="border-border/50 bg-background rounded-lg border px-3 py-2 text-xs shadow-xl">
                            <div className="font-medium">{point.hour} Uhr</div>
                            <div className="text-muted-foreground">{point.value}× fertiggestellt</div>
                          </div>
                        )
                      }}
                    />
                    <Bar dataKey="value" fill="var(--color-value)" radius={[4, 4, 0, 0]} />
                  </BarChart>
                </ChartContainer>
              )}
            </CardContent>
          </Card>

          <TopProductsChart title="Top-Produkte der letzten Stunde" data={topProductsData} loading={loading} />
        </div>
      </div>
    </div>
  )
}
