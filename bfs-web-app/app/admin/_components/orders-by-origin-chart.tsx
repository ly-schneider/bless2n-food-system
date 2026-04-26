"use client"

import { Cell, Pie, PieChart } from "recharts"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { type ChartConfig, ChartContainer, ChartTooltip } from "@/components/ui/chart"

type OriginData = { origin: string; label: string; count: number }

const COLORS = ["var(--chart-1)", "var(--chart-3)"]

const chartConfig = {
  count: { label: "Anzahl" },
} satisfies ChartConfig

export function OrdersByOriginChart({ data, loading }: { data: OriginData[]; loading: boolean }) {
  const total = data.reduce((s, d) => s + d.count, 0)

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base">Herkunft</CardTitle>
      </CardHeader>
      <CardContent>
        {loading ? (
          <div className="text-muted-foreground flex h-[250px] items-center justify-center text-sm">Lade…</div>
        ) : data.length === 0 ? (
          <div className="text-muted-foreground flex h-[250px] items-center justify-center text-sm">
            Keine Daten vorhanden.
          </div>
        ) : (
          <div className="flex flex-col items-center">
            <ChartContainer config={chartConfig} className="h-[200px] w-full">
              <PieChart>
                <ChartTooltip
                  content={({ active, payload }) => {
                    if (!active || !payload?.length || !payload[0]) return null
                    const d = payload[0].payload as OriginData
                    const pct = total > 0 ? ((d.count / total) * 100).toFixed(1) : "0"
                    return (
                      <div className="border-border/50 bg-background rounded-lg border px-3 py-2 text-xs shadow-xl">
                        <div className="font-medium">{d.label}</div>
                        <div className="text-muted-foreground">
                          {d.count}× ({pct}%)
                        </div>
                      </div>
                    )
                  }}
                />
                <Pie data={data} dataKey="count" nameKey="label" innerRadius={50} outerRadius={80} strokeWidth={2}>
                  {data.map((_, i) => (
                    <Cell key={i} fill={COLORS[i % COLORS.length]} />
                  ))}
                </Pie>
              </PieChart>
            </ChartContainer>
            <div className="mt-2 flex flex-wrap justify-center gap-x-4 gap-y-1 text-xs">
              {data.map((d, i) => (
                <div key={d.origin} className="flex items-center gap-1.5">
                  <div className="size-2 rounded-full" style={{ backgroundColor: COLORS[i % COLORS.length] }} />
                  <span className="text-muted-foreground">{d.label}</span>
                </div>
              ))}
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  )
}
