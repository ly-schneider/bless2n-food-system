"use client"

import { Bar, BarChart, XAxis, YAxis } from "recharts"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { type ChartConfig, ChartContainer, ChartTooltip } from "@/components/ui/chart"
import { formatChf } from "@/lib/utils"

type ProductData = { title: string; value: number }

type TopProductsChartProps = {
  title: string
  data: ProductData[]
  loading: boolean
  formatAsCurrency?: boolean
}

export function TopProductsChart({ title, data, loading, formatAsCurrency }: TopProductsChartProps) {
  const chartConfig = {
    value: {
      label: formatAsCurrency ? "Umsatz" : "Menge",
      color: formatAsCurrency ? "var(--chart-3)" : "var(--chart-2)",
    },
  } satisfies ChartConfig

  const chartHeight = Math.max(200, data.length * 32 + 40)

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base">{title}</CardTitle>
      </CardHeader>
      <CardContent>
        {loading ? (
          <div className="text-muted-foreground flex h-[250px] items-center justify-center text-sm">Lade…</div>
        ) : data.length === 0 ? (
          <div className="text-muted-foreground flex h-[250px] items-center justify-center text-sm">
            Keine Daten vorhanden.
          </div>
        ) : (
          <ChartContainer config={chartConfig} className="w-full" style={{ height: chartHeight }}>
            <BarChart data={data} layout="vertical" margin={{ top: 0, right: 4, bottom: 0, left: 0 }}>
              <XAxis
                type="number"
                tickLine={false}
                axisLine={false}
                fontSize={12}
                tickFormatter={formatAsCurrency ? (v: number) => `${Math.round(v / 100)}` : undefined}
              />
              <YAxis
                type="category"
                dataKey="title"
                tickLine={false}
                axisLine={false}
                fontSize={12}
                width={120}
                tickFormatter={(v: string) => (v.length > 16 ? v.slice(0, 14) + "…" : v)}
              />
              <ChartTooltip
                content={({ active, payload }) => {
                  if (!active || !payload?.length || !payload[0]) return null
                  const d = payload[0].payload as ProductData
                  return (
                    <div className="border-border/50 bg-background rounded-lg border px-3 py-2 text-xs shadow-xl">
                      <div className="font-medium">{d.title}</div>
                      <div className="text-muted-foreground">
                        {formatAsCurrency ? formatChf(d.value) : `${d.value}×`}
                      </div>
                    </div>
                  )
                }}
              />
              <Bar dataKey="value" fill="var(--color-value)" radius={[0, 4, 4, 0]} />
            </BarChart>
          </ChartContainer>
        )}
      </CardContent>
    </Card>
  )
}
