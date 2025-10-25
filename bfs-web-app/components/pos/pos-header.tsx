"use client"
import { Delete, Lock, RefreshCw } from "lucide-react"
import Image from "next/image"
import { useEffect, useRef, useState } from "react"
import { createPortal } from "react-dom"
import { PrinterSelector } from "@/components/pos/printer-selector"
import { Button } from "@/components/ui/button"
import { InputOTP, InputOTPGroup, InputOTPSlot } from "@/components/ui/input-otp"

export function POSHeader() {
  const PIN = process.env.NEXT_PUBLIC_POS_PIN || "0000"
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
  const [pin, setPin] = useState("")
  const [error, setError] = useState<string | null>(null)
  const idleTimerRef = useRef<number | null>(null)
  const [mounted, setMounted] = useState(false)

  function lock() {
    setPin("")
    setError(null)
    setLocked(true)
    try {
      window.dispatchEvent(new CustomEvent("pos:lock"))
    } catch {}
  }
  function tryUnlock(candidate?: string) {
    const value = candidate ?? pin
    if (value === PIN) {
      setLocked(false)
      setPin("")
      setError(null)
      try {
        window.dispatchEvent(new CustomEvent("pos:unlock"))
      } catch {}
    } else {
      setError("Falscher PIN")
    }
  }

  // Persist lock state across refreshes
  useEffect(() => {
    try {
      localStorage.setItem(LOCK_KEY, locked ? "1" : "0")
    } catch {}
  }, [locked])

  // Mark as mounted for safe portal usage
  useEffect(() => {
    setMounted(true)
  }, [])

  // Idle auto-lock handling
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
    // Start/restart timer when unlocked
    if (!locked) resetIdle()
    // Global activity listeners
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

  function onOtpChange(v: string) {
    const val = v.replace(/\D/g, "").slice(0, 4)
    setPin(val)
    setError(null)
    if (val.length === 4) {
      console.log("Checking PIN", val, PIN)
      if (val === PIN) {
        tryUnlock(val)
      } else {
        setError("Falscher PIN")
        setTimeout(() => setPin(""), 150)
      }
    }
  }

  function onKeyDigit(d: string) {
    if (pin.length >= 4) return
    onOtpChange(pin + d)
  }
  function onBackspace() {
    if (pin.length === 0) return
    onOtpChange(pin.slice(0, -1))
  }
  function onClear() {
    onOtpChange("")
  }

  const overlay = locked ? (
    <div className="bg-background fixed inset-0 z-[10000] grid min-h-dvh w-dvw place-items-center">
      <div className="w-full max-w-sm">
        <div className="mb-6 flex items-center justify-center gap-2">
          <div className="bg-muted flex h-12 w-12 items-center justify-center overflow-hidden rounded-full">
            <Image src="/assets/images/blessthun.png" alt="BlessThun" width={48} height={48} />
          </div>
          <span className="text-lg font-semibold">BlessThun</span>
        </div>
        <h2 className="mb-10 text-center text-5xl font-semibold">Gesperrt</h2>
        <div className="grid gap-10">
          <div>
            {/* Visual masked indicators for 4-digit PIN */}
            <div className="mt-3 flex items-center justify-center gap-3">
              {[0, 1, 2, 3].map((i) => (
                <div
                  key={i}
                  aria-hidden
                  className={
                    (i < pin.length ? "bg-foreground" : "bg-muted-foreground/30") +
                    " h-4 w-4 rounded-full"
                  }
                />
              ))}
            </div>
            {error && <div className="text-destructive mt-10 text-sm text-center">{error}</div>}
          </div>
          {/* Numeric keypad */}
          <div className="mt-2 grid grid-cols-3 gap-2">
            {["1", "2", "3", "4", "5", "6", "7", "8", "9"].map((d) => (
              <Button key={d} variant="outline" onClick={() => onKeyDigit(d)} className="h-16 text-lg">
                {/* Hide actual digit label, show generic bullet */}
                <span aria-hidden className="select-none text-2xl">{d}</span>
                <span className="sr-only">Ziffer</span>
              </Button>
            ))}
            <Button variant="outline" onClick={onClear} className="h-16 text-lg">
              C
            </Button>
            <Button variant="outline" onClick={() => onKeyDigit("0")} className="h-16 text-lg">
              <span aria-hidden className="select-none text-2xl">0</span>
              <span className="sr-only">Ziffer</span>
            </Button>
            <Button variant="outline" onClick={onBackspace} className="h-16 text-lg">
              <Delete className="size-4" />
            </Button>
          </div>
        </div>
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
            <PrinterSelector />
            <Button
              variant="outline"
              size="sm"
              onClick={() => {
                try {
                  window.dispatchEvent(new CustomEvent("admin:refresh"))
                } catch {}
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
