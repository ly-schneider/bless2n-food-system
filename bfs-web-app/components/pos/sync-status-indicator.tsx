"use client"

import { AlertCircle, CloudOff, Loader2 } from "lucide-react"
import { Button } from "@/components/ui/button"

interface SyncStatusIndicatorProps {
  isOnline: boolean
  pendingCount: number
  failedCount: number
  onFailedClick?: () => void
}

export function SyncStatusIndicator({ isOnline, pendingCount, failedCount, onFailedClick }: SyncStatusIndicatorProps) {
  if (isOnline && pendingCount === 0 && failedCount === 0) {
    return null
  }

  return (
    <div className="flex items-center gap-1.5">
      {!isOnline && (
        <Button
          variant="outline"
          size="sm"
          className="bg-destructive/10 text-destructive pointer-events-none rounded-[11px] border-0"
        >
          <CloudOff className="size-4" />
          <span>Offline</span>
        </Button>
      )}

      {pendingCount > 0 && (
        <Button variant="outline" size="sm" className="pointer-events-none rounded-[11px] border-0">
          <Loader2 className="size-4 animate-spin" />
          <span>{pendingCount}</span>
        </Button>
      )}

      {failedCount > 0 && (
        <Button
          variant="outline"
          size="sm"
          className="bg-destructive/10 text-destructive hover:bg-destructive/20 hover:text-destructive rounded-[11px] border-0"
          onClick={onFailedClick}
        >
          <AlertCircle className="size-4" />
          <span>{failedCount}</span>
        </Button>
      )}
    </div>
  )
}
