"use client"
import { Lock, RefreshCw } from "lucide-react"
import Image from "next/image"
import { useCallback, useEffect, useRef, useState } from "react"
import { createPortal } from "react-dom"
import { PrinterSelector } from "@/components/pos/printer-selector"
import { SyncStatusIndicator } from "@/components/pos/sync-status-indicator"
import { Button } from "@/components/ui/button"
import type { PosFulfillmentMode } from "@/types/jeton"

interface POSHeaderProps {
  mode?: PosFulfillmentMode
  syncStatus?: {
    isOnline: boolean
    pendingCount: number
    failedCount: number
    onFailedClick?: () => void
  }
}

const HOLD_DURATION_MS = 1000

export function POSHeader({ mode, syncStatus }: POSHeaderProps) {
  const LOCK_KEY = "bfs.pos.locked"
  const IDLE_MS = Number(process.env.NEXT_PUBLIC_POS_IDLE_TIMEOUT) || 300000
  const [locked, setLocked] = useState<boolean>(() => {
    if (typeof window === "undefined") return false
    try {
      return localStorage.getItem(LOCK_KEY) === "1"
    } catch {
      return false
    }
  })
  const [holdProgress, setHoldProgress] = useState(0)
  const holdStartRef = useRef<number | null>(null)
  const animationFrameRef = useRef<number | null>(null)
  const idleTimerRef = useRef<number | null>(null)
  const [mounted, setMounted] = useState(false)
  const jetonMode = mode === "JETON"

  function lock() {
    setLocked(true)
    try {
      window.dispatchEvent(new CustomEvent("pos:lock"))
    } catch {}
  }

  function unlock() {
    setLocked(false)
    setHoldProgress(0)
    holdStartRef.current = null
    try {
      window.dispatchEvent(new CustomEvent("pos:unlock"))
    } catch {}
  }

  const cancelHold = useCallback(() => {
    holdStartRef.current = null
    if (animationFrameRef.current) {
      cancelAnimationFrame(animationFrameRef.current)
      animationFrameRef.current = null
    }
    setHoldProgress(0)
  }, [])

  const updateProgress = useCallback(() => {
    if (holdStartRef.current === null) return

    const elapsed = performance.now() - holdStartRef.current
    const progress = Math.min(elapsed / HOLD_DURATION_MS, 1)
    setHoldProgress(progress)

    if (progress >= 1) {
      unlock()
    } else {
      animationFrameRef.current = requestAnimationFrame(updateProgress)
    }
  }, [])

  const startHold = useCallback(() => {
    holdStartRef.current = performance.now()
    animationFrameRef.current = requestAnimationFrame(updateProgress)
  }, [updateProgress])

  useEffect(() => {
    try {
      localStorage.setItem(LOCK_KEY, locked ? "1" : "0")
    } catch {}
  }, [locked])

  useEffect(() => {
    setMounted(true)
  }, [])

  useEffect(() => {
    function resetIdle() {
      if (idleTimerRef.current != null) window.clearTimeout(idleTimerRef.current)
      idleTimerRef.current = window.setTimeout(() => {
        setLocked(true)
      }, IDLE_MS)
    }
    function onActivity() {
      if (!locked) resetIdle()
    }
    if (!locked) resetIdle()
    window.addEventListener("pointerdown", onActivity)
    window.addEventListener("keydown", onActivity)
    window.addEventListener("touchstart", onActivity, { passive: true })
    window.addEventListener("mousemove", onActivity)
    return () => {
      window.removeEventListener("pointerdown", onActivity)
      window.removeEventListener("keydown", onActivity)
      window.removeEventListener("touchstart", onActivity)
      window.removeEventListener("mousemove", onActivity)
      if (idleTimerRef.current != null) window.clearTimeout(idleTimerRef.current)
    }
  }, [locked])

  useEffect(() => {
    return () => {
      if (animationFrameRef.current) {
        cancelAnimationFrame(animationFrameRef.current)
      }
    }
  }, [])

  const [viewportSize, setViewportSize] = useState({ w: 0, h: 0 })
  useEffect(() => {
    const update = () => setViewportSize({ w: window.innerWidth, h: window.innerHeight })
    update()
    window.addEventListener("resize", update)
    return () => window.removeEventListener("resize", update)
  }, [])

  const circleSize = holdProgress * Math.max(viewportSize.w, viewportSize.h) * 1.5

  const overlay = locked ? (
    <div
      className="bg-background fixed inset-0 z-[10000] flex min-h-dvh w-dvw cursor-pointer items-center justify-center overflow-hidden select-none"
      onPointerDown={startHold}
      onPointerUp={cancelHold}
      onPointerLeave={cancelHold}
      onPointerCancel={cancelHold}
      onContextMenu={(e) => e.preventDefault()}
    >
      {holdProgress > 0 && (
        <div
          className="bg-muted pointer-events-none absolute rounded-full"
          style={{
            width: circleSize,
            height: circleSize,
            transform: "translate(-50%, -50%)",
            left: "50%",
            top: "50%",
          }}
        />
      )}
      <div className="pointer-events-none relative z-10 flex flex-col items-center gap-4">
        <div className="mb-2 flex items-center gap-2">
          <div className="bg-muted flex h-10 w-10 items-center justify-center overflow-hidden rounded-full mr-1">
            <Image src="/assets/images/blessthun.png" alt="BlessThun" width={40} height={40} />
          </div>
          <span className="text-lg font-semibold">BlessThun</span>
        </div>
        <h2 className="text-6xl font-bold tracking-tight">GESPERRT</h2>
      </div>
    </div>
  ) : null

  return (
    <div className="bg-background sticky top-0 z-[50]">
      <div className="mx-auto w-full p-3 md:p-4">
        <div className="flex h-8 items-center justify-between">
          <div className="flex items-center gap-2">
            <div className="bg-muted flex h-9 w-9 items-center justify-center overflow-hidden rounded-full">
              <Image src="/assets/images/blessthun.png" alt="BlessThun" width={36} height={36} />
            </div>
            <span className="text-sm font-semibold">BlessThun</span>
          </div>
          <div className="flex items-center gap-1.5">
            {syncStatus && (
              <SyncStatusIndicator
                isOnline={syncStatus.isOnline}
                pendingCount={syncStatus.pendingCount}
                failedCount={syncStatus.failedCount}
                onFailedClick={syncStatus.onFailedClick}
              />
            )}
            {!jetonMode && <PrinterSelector />}
            <Button
              variant="outline"
              size="sm"
              onClick={() => {
                const url = new URL(window.location.href)
                url.searchParams.set("_r", Date.now().toString())
                window.location.href = url.toString()
              }}
              aria-label="Aktualisieren"
              className="rounded-[11px] border-0"
            >
              <RefreshCw className="size-4" />
              <span>Aktualisieren</span>
            </Button>
            <Button
              variant="outline"
              size="icon"
              aria-label="Sperren"
              className="rounded-[11px] border-0"
              onClick={lock}
            >
              <Lock className="size-4" />
            </Button>
          </div>
        </div>
      </div>
      {mounted ? createPortal(overlay, document.body) : overlay}
    </div>
  )
}
