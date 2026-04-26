"use client"

import { Bar, BarChart, CartesianGrid, XAxis, YAxis } from "recharts"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { type ChartConfig, ChartContainer, ChartTooltip } from "@/components/ui/chart"
import { formatChf } from "@/lib/utils"

type RevenueByHourData = { hour: string; revenueCents: number }

const chartConfig = {
  revenueCents: { label: "Umsatz", color: "var(--chart-1)" },
} satisfies ChartConfig

export function RevenueByHourChart({ data, loading }: { data: RevenueByHourData[]; loading: boolean }) {
  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base">Umsatz pro Stunde</CardTitle>
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
              <YAxis
                tickLine={false}
                axisLine={false}
                fontSize={12}
                tickFormatter={(v: number) => `${Math.round(v / 100)}`}
                width={40}
              />
              <ChartTooltip
                content={({ active, payload }) => {
                  if (!active || !payload?.length || !payload[0]) return null
                  const d = payload[0].payload as RevenueByHourData
                  return (
                    <div className="border-border/50 bg-background rounded-lg border px-3 py-2 text-xs shadow-xl">
                      <div className="font-medium">{d.hour} Uhr</div>
                      <div className="text-muted-foreground">{formatChf(d.revenueCents)}</div>
                    </div>
                  )
                }}
              />
              <Bar dataKey="revenueCents" fill="var(--color-revenueCents)" radius={[4, 4, 0, 0]} />
            </BarChart>
          </ChartContainer>
        )}
      </CardContent>
    </Card>
  )
}
