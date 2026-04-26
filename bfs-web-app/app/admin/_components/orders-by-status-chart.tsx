"use client"

import { Bar, BarChart, CartesianGrid, XAxis, YAxis } from "recharts"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { type ChartConfig, ChartContainer, ChartLegend, ChartLegendContent, ChartTooltip } from "@/components/ui/chart"

type StatusByHourData = {
  hour: string
  paid: number
  pending: number
  cancelled: number
  refunded: number
}

const chartConfig = {
  paid: { label: "Bezahlt", color: "hsl(142 71% 45%)" },
  pending: { label: "Ausstehend", color: "hsl(217 91% 60%)" },
  cancelled: { label: "Storniert", color: "hsl(25 95% 53%)" },
  refunded: { label: "Erstattet", color: "hsl(0 84% 60%)" },
} satisfies ChartConfig

export function OrdersByStatusChart({ data, loading }: { data: StatusByHourData[]; loading: boolean }) {
  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base">Bestellungen nach Status</CardTitle>
      </CardHeader>
      <CardContent>
        {loading ? (
          <div className="text-muted-foreground flex h-[250px] items-center justify-center text-sm">Lade…</div>
        ) : data.length === 0 ? (
          <div className="text-muted-foreground flex h-[250px] items-center justify-center text-sm">
            Keine Daten vorhanden.
          </div>
        ) : (
          <ChartContainer config={chartConfig} className="h-[250px] w-full">
            <BarChart data={data} margin={{ top: 4, right: 4, bottom: 0, left: 0 }}>
              <CartesianGrid vertical={false} strokeDasharray="3 3" />
              <XAxis dataKey="hour" tickLine={false} axisLine={false} fontSize={12} />
              <YAxis tickLine={false} axisLine={false} fontSize={12} width={30} allowDecimals={false} />
              <ChartTooltip
                content={({ active, payload, label }) => {
                  if (!active || !payload?.length) return null
                  return (
                    <div className="border-border/50 bg-background rounded-lg border px-3 py-2 text-xs shadow-xl">
                      <div className="mb-1 font-medium">{label} Uhr</div>
                      {payload.map((p) => {
                        const key = String(p.dataKey) as keyof typeof chartConfig
                        const cfg = chartConfig[key]
                        return (
                          <div key={key} className="flex items-center gap-2">
                            <div className="size-2 rounded-full" style={{ backgroundColor: cfg?.color }} />
                            <span className="text-muted-foreground">{cfg?.label}:</span>
                            <span className="font-medium">{p.value}</span>
                          </div>
                        )
                      })}
                    </div>
                  )
                }}
              />
              <ChartLegend content={<ChartLegendContent />} />
              <Bar dataKey="paid" stackId="status" fill="var(--color-paid)" />
              <Bar dataKey="pending" stackId="status" fill="var(--color-pending)" />
              <Bar dataKey="cancelled" stackId="status" fill="var(--color-cancelled)" />
              <Bar dataKey="refunded" stackId="status" fill="var(--color-refunded)" radius={[4, 4, 0, 0]} />
            </BarChart>
          </ChartContainer>
        )}
      </CardContent>
    </Card>
  )
}
