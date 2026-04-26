import { TrendingDown, TrendingUp } from "lucide-react"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { cn } from "@/lib/utils"

type StatCardProps = {
  title: string
  value: string
  loading: boolean
  comparison?: { label: string; positive: boolean } | null
}

export function StatCard({ title, value, loading, comparison }: StatCardProps) {
  return (
    <Card className="gap-2 py-4">
      <CardHeader className="pb-0">
        <CardTitle className="text-muted-foreground text-sm font-medium">{title}</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="text-3xl font-semibold">{loading ? "–" : value}</div>
        {comparison && !loading && (
          <div
            className={cn(
              "mt-1.5 flex items-center gap-1 text-xs font-medium",
              comparison.positive ? "text-green-600" : "text-red-600"
            )}
          >
            {comparison.positive ? <TrendingUp className="size-3.5" /> : <TrendingDown className="size-3.5" />}
            {comparison.label}
          </div>
        )}
      </CardContent>
    </Card>
  )
}
