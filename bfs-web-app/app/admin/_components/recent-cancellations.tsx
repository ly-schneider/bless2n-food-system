"use client"

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { cn, formatChf } from "@/lib/utils"

type CancellationItem = {
  id: string
  status: string
  totalCents: number
  origin: string
  createdAt: string
}

const STATUS_LABELS: Record<string, string> = {
  cancelled: "Storniert",
  refunded: "Erstattet",
}

const ORIGIN_LABELS: Record<string, string> = {
  shop: "Shop",
  pos: "POS",
}

export function RecentCancellations({ data, loading }: { data: CancellationItem[]; loading: boolean }) {
  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base">Letzte Stornierungen</CardTitle>
      </CardHeader>
      <CardContent>
        {loading ? (
          <div className="text-muted-foreground flex h-[250px] items-center justify-center text-sm">Lade…</div>
        ) : data.length === 0 ? (
          <div className="text-muted-foreground flex h-[250px] items-center justify-center text-sm">
            Keine Stornierungen.
          </div>
        ) : (
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Zeit</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Betrag</TableHead>
                <TableHead>Herkunft</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {data.map((item) => (
                <TableRow key={item.id}>
                  <TableCell className="text-muted-foreground">
                    {new Date(item.createdAt).toLocaleTimeString("de-CH", {
                      hour: "2-digit",
                      minute: "2-digit",
                    })}
                  </TableCell>
                  <TableCell>
                    <span
                      className={cn(
                        "inline-flex rounded-full px-2 py-0.5 text-xs font-medium",
                        item.status === "cancelled" ? "bg-orange-100 text-orange-700" : "bg-red-100 text-red-700"
                      )}
                    >
                      {STATUS_LABELS[item.status] ?? item.status}
                    </span>
                  </TableCell>
                  <TableCell className="font-medium">{formatChf(item.totalCents)}</TableCell>
                  <TableCell className="text-muted-foreground">{ORIGIN_LABELS[item.origin] ?? item.origin}</TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        )}
      </CardContent>
    </Card>
  )
}
