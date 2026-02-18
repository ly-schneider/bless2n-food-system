"use client"

import { AlertCircle, RefreshCw, Trash2 } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import { formatChf } from "@/lib/utils"
import type { QueuedOrder } from "@/types/order-queue"

interface StuckOrdersDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  orders: QueuedOrder[]
  onRetry: (localId: string) => void
  onDelete: (localId: string) => void
}

function formatTime(timestamp: number): string {
  return new Date(timestamp).toLocaleTimeString("de-CH", {
    hour: "2-digit",
    minute: "2-digit",
  })
}

export function StuckOrdersDialog({ open, onOpenChange, orders, onRetry, onDelete }: StuckOrdersDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <AlertCircle className="text-destructive size-5" />
            Fehlgeschlagene Bestellungen
          </DialogTitle>
        </DialogHeader>

        <div className="max-h-[60vh] space-y-3 overflow-y-auto">
          {orders.length === 0 ? (
            <p className="text-muted-foreground py-4 text-center text-sm">Keine fehlgeschlagenen Bestellungen</p>
          ) : (
            orders.map((order) => (
              <div key={order.localId} className="bg-muted/50 flex items-start justify-between rounded-lg border p-3">
                <div className="min-w-0 flex-1">
                  <div className="flex items-center gap-2">
                    <span className="text-sm font-medium">{formatChf(order.totalCents)}</span>
                    <span className="text-muted-foreground text-xs">{formatTime(order.createdAt)}</span>
                  </div>
                  <div className="text-muted-foreground mt-0.5 text-xs">
                    {order.items.length} {order.items.length === 1 ? "Produkt" : "Produkte"} &middot;{" "}
                    {order.paymentMethod === "cash" ? "Bar" : order.paymentMethod === "card" ? "Karte" : "TWINT"}
                  </div>
                  {order.lastError && <div className="text-destructive mt-1 truncate text-xs">{order.lastError}</div>}
                </div>
                <div className="ml-3 flex shrink-0 gap-1">
                  <Button
                    variant="ghost"
                    size="icon"
                    className="size-8"
                    onClick={() => onRetry(order.localId)}
                    title="Erneut versuchen"
                  >
                    <RefreshCw className="size-4" />
                  </Button>
                  <Button
                    variant="ghost"
                    size="icon"
                    className="text-destructive hover:text-destructive size-8"
                    onClick={() => onDelete(order.localId)}
                    title="Löschen"
                  >
                    <Trash2 className="size-4" />
                  </Button>
                </div>
              </div>
            ))
          )}
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Schliessen
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
