"use client"

import { useCallback, useEffect, useRef, useState } from "react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { getClientInfo } from "@/lib/client-info"
import { randomUrlSafe } from "@/lib/crypto"
import { setDeviceToken, setDeviceType } from "@/lib/device-auth"

type DeviceType = "POS" | "STATION"

interface PairingCodeDisplayProps {
  deviceType: DeviceType
  defaultName?: string
}

type PairingState = { phase: "input" } | { phase: "showing"; code: string; expiresAt: Date } | { phase: "expired" }

type PollResponse = {
  status: "pending" | "completed" | "expired"
  token?: string
  device?: { id?: string; name?: string }
}

const POLL_INTERVAL_MS = 4000

function formatCode(code: string): string {
  if (code.length <= 3) return code
  return `${code.slice(0, 3)} ${code.slice(3)}`
}

export function PairingCodeDisplay({ deviceType, defaultName }: PairingCodeDisplayProps) {
  const [name, setName] = useState(defaultName ?? "")
  const [state, setState] = useState<PairingState>({ phase: "input" })
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState(false)
  const pollingRef = useRef(false)
  const timerRef = useRef<ReturnType<typeof setInterval> | null>(null)
  const [countdown, setCountdown] = useState(0)

  // Generate a stable device key per browser
  const deviceKey = useRef<string>("")
  if (typeof window !== "undefined" && !deviceKey.current) {
    const storageKey = `bfs.pairingDeviceKey.${deviceType.toLowerCase()}`
    let k = localStorage.getItem(storageKey)
    if (!k) {
      k = `dev_${randomUrlSafe(24)}`
      localStorage.setItem(storageKey, k)
    }
    deviceKey.current = k
  }

  const generateCode = useCallback(async () => {
    setError(null)
    setLoading(true)
    try {
      const info = await getClientInfo()
      const res = await fetch(`/api/v1/devices/pairings`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          name: name || (deviceType === "POS" ? "POS-Terminal" : "Station"),
          model: info.model,
          os: info.os,
          deviceKey: deviceKey.current,
          type: deviceType,
        }),
      })
      if (!res.ok) {
        throw new Error(`HTTP ${res.status}`)
      }
      const data = (await res.json()) as { code: string; expiresAt: string }
      setState({ phase: "showing", code: data.code, expiresAt: new Date(data.expiresAt) })
    } catch {
      setError("Code konnte nicht generiert werden.")
    } finally {
      setLoading(false)
    }
  }, [name, deviceType])

  // Countdown timer
  useEffect(() => {
    if (state.phase !== "showing") return
    const update = () => {
      const remaining = Math.max(0, Math.floor((state.expiresAt.getTime() - Date.now()) / 1000))
      setCountdown(remaining)
      if (remaining <= 0) {
        setState({ phase: "expired" })
        pollingRef.current = false
      }
    }
    update()
    const id = setInterval(update, 1000)
    return () => clearInterval(id)
  }, [state])

  // Polling
  useEffect(() => {
    if (state.phase !== "showing") {
      pollingRef.current = false
      return
    }

    pollingRef.current = true

    const poll = async () => {
      if (!pollingRef.current) return
      try {
        const res = await fetch(`/api/v1/devices/pairings/${encodeURIComponent(state.code)}`)
        if (!res.ok) return
        const data = (await res.json()) as PollResponse
        if (data.status === "completed" && data.token) {
          pollingRef.current = false
          setDeviceToken(data.token)
          setDeviceType(deviceType.toLowerCase() as "pos" | "station")
          window.location.reload()
        } else if (data.status === "expired") {
          pollingRef.current = false
          setState({ phase: "expired" })
        }
      } catch {
        // Retry on next interval
      }
    }

    poll()
    timerRef.current = setInterval(poll, POLL_INTERVAL_MS)
    return () => {
      pollingRef.current = false
      if (timerRef.current) clearInterval(timerRef.current)
    }
  }, [state, deviceType])

  const placeholderName = deviceType === "POS" ? "z. B. Kasse 1" : "z. B. Grill 1"
  const title = deviceType === "POS" ? "POS-Zugang anfordern" : "Station registrieren"

  return (
    <div className="grid min-h-[calc(100dvh-8rem)] place-items-center p-4">
      <div className="bg-background w-full max-w-md rounded-xl border p-5 shadow-sm">
        <h1 className="mb-2 text-2xl font-semibold">{title}</h1>
        <p className="text-muted-foreground mb-4 text-sm">Dieses Gerät muss von einem Admin gekoppelt werden.</p>

        {error && (
          <div className="mb-3 rounded-xl border border-red-200 bg-red-50 p-2 text-sm text-red-700">{error}</div>
        )}

        {state.phase === "input" && (
          <div className="grid gap-3">
            <div className="grid gap-1">
              <Label htmlFor="device-name">Gerätename</Label>
              <Input
                id="device-name"
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder={placeholderName}
              />
            </div>
            <Button onClick={generateCode} disabled={loading}>
              {loading ? "Wird generiert..." : "Code generieren"}
            </Button>
          </div>
        )}

        {state.phase === "showing" && (
          <div className="grid gap-4">
            <div className="rounded-xl border bg-gray-50 p-6 text-center dark:bg-gray-900">
              <p className="text-muted-foreground mb-1 text-xs">Kopplungscode</p>
              <p className="my-4 text-4xl font-bold">{formatCode(state.code)}</p>
              <p className="text-muted-foreground mt-4 text-xs">
                {countdown > 0
                  ? `Gültig für ${Math.floor(countdown / 60)}:${String(countdown % 60).padStart(2, "0")}`
                  : "Abgelaufen"}
              </p>
            </div>
            <div className="flex items-center gap-2 rounded-xl border border-amber-200 bg-amber-50 p-3 text-sm text-amber-700">
              <svg className="h-4 w-4 shrink-0 animate-spin" viewBox="0 0 24 24" fill="none">
                <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
                <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
              </svg>
              Gib diesen Code im Admin-Dashboard ein.
            </div>
          </div>
        )}

        {state.phase === "expired" && (
          <div className="grid gap-3">
            <div className="rounded-xl border border-amber-200 bg-amber-50 p-3 text-sm text-amber-700">
              Code abgelaufen.
            </div>
            <Button onClick={generateCode} disabled={loading}>
              {loading ? "Wird generiert..." : "Neuen Code generieren"}
            </Button>
          </div>
        )}
      </div>
    </div>
  )
}
