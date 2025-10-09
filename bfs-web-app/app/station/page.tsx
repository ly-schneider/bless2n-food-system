"use client"

import { useEffect, useMemo, useRef, useState } from "react"
import { API_BASE_URL } from "@/lib/api"
import Image from "next/image"
import {
  Dialog,
  DialogContent,
  DialogFooter as ModalFooter,
  DialogHeader as ModalHeader,
  DialogTitle as ModalTitle,
} from "@/components/ui/dialog"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { Button } from "@/components/ui/button"

type StationStatus = { exists: boolean; approved: boolean; name?: string }

function randKey(): string {
  const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_"
  const bytes = new Uint8Array(24)
  if (typeof crypto !== "undefined" && crypto.getRandomValues) crypto.getRandomValues(bytes)
  return Array.from(bytes, (b) => chars[b % chars.length]).join("")
}

export default function StationPage() {
  const log = (...args: any[]) => console.log('[Station]', ...args)
  const videoRef = useRef<HTMLVideoElement | null>(null)
  const [devices, setDevices] = useState<MediaDeviceInfo[]>([])
  const [deviceId, setDeviceId] = useState<string>("")
  const [status, setStatus] = useState<StationStatus | null>(null)
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState<string | null>(null)
  type PublicOrderItem = {
    id: string
    orderId: string
    productId: string
    title: string
    quantity: number
    pricePerUnitCents?: number
    productImage?: string | null
    isRedeemed?: boolean
    parentItemId?: string | null
    menuSlotId?: string | null
    menuSlotName?: string | null
  }
  type VerifyResult = { orderId: string; items: PublicOrderItem[] }
  const [result, setResult] = useState<VerifyResult | null>(null)
  const [scanned, setScanned] = useState<string | null>(null)
  // Redemption response is not shown; keep UI minimal
  const [stationName, setStationName] = useState("")
  const [drawerOpen, setDrawerOpen] = useState(false)
  const [requestSubmitted, setRequestSubmitted] = useState(false)
  // ZXing removed; using native BarcodeDetector
  const [scanningPaused, setScanningPaused] = useState(false)
  const pausedRef = useRef(false)
  const detectorRef = useRef<any | null>(null)
  const rafRef = useRef<number | null>(null)
  const lastCodeRef = useRef<string | null>(null)
  const lastAtRef = useRef<number>(0)
  useEffect(() => {
    pausedRef.current = scanningPaused || drawerOpen
  }, [scanningPaused, drawerOpen])

  const stationKey = useMemo(() => {
    if (typeof window === "undefined") return ""
    let k = localStorage.getItem("bfs.stationKey")
    if (!k) {
      k = `st_${randKey()}`
      localStorage.setItem("bfs.stationKey", k)
    }
    log('stationKey', k)
    return k
  }, [])
  // No offline mode or service worker – simplify flow

  useEffect(() => {
    // fetch station status
    ;(async () => {
      try {
        log('GET /v1/stations/me')
        const res = await fetch(`${API_BASE_URL}/v1/stations/me`, { headers: { "X-Station-Key": stationKey } })
        const json = await res.json()
        setStatus(json as StationStatus)
        log('status', json)
      } catch {
        setStatus({ exists: false, approved: false })
        log('status fetch failed')
      }
    })()
  }, [stationKey])

  // After approval: request camera permission, enumerate devices, and start scanning
  useEffect(() => {
    if (!status?.approved) return
    ;(async () => {
      try {
        // Request permission with default camera to reveal labels
        log('Requesting camera permission')
        await navigator.mediaDevices.getUserMedia({ video: true })
        log('Camera permission granted')
        const ds = await navigator.mediaDevices.enumerateDevices()
        const vids = ds.filter((d) => d.kind === "videoinput")
        setDevices(vids)
        log('Cameras', vids.map(v => ({ id: v.deviceId, label: v.label })))
        const preferred = vids[0]?.deviceId || ""
        setDeviceId((prev) => prev || preferred)
      } catch (e) {
        setError("Kamerazugriff verweigert oder nicht verfügbar.")
        log('Camera permission failed', e)
      }
    })()
  }, [status?.approved])

  // Cleanup on unmount: stop camera tracks and cancel RAF
  useEffect(() => {
    return () => {
      try {
        const v = videoRef.current
        const tracks = (v?.srcObject as MediaStream | null)?.getTracks() || []
        tracks.forEach((t) => t.stop())
      } catch {}
      if (rafRef.current) cancelAnimationFrame(rafRef.current)
    }
  }, [])

  // Native BarcodeDetector scanning loop (camera stays active)
  useEffect(() => {
    if (!status?.approved) return
    const BD = (typeof window !== 'undefined' ? (window as any).BarcodeDetector : null)
    if (!BD) return
    if (!detectorRef.current) {
      try { detectorRef.current = new BD({ formats: ['qr_code'] }); log('BarcodeDetector initialized') } catch { return }
    }
    let cancelled = false
    const tick = async () => {
      if (cancelled) return
      if (!(scanningPaused || drawerOpen) && videoRef.current && detectorRef.current) {
        try {
          const codes = await detectorRef.current.detect(videoRef.current)
          const found = codes?.find((c: any) => c.rawValue)
          if (found?.rawValue) {
            const now = Date.now()
            const value = String(found.rawValue)
            const isRepeat = lastCodeRef.current === value && (now - lastAtRef.current) < 1500
            if (!isRepeat) {
              lastCodeRef.current = value
              lastAtRef.current = now
              setScanningPaused(true)
              log('QR detected', value.slice(0, 24) + (value.length > 24 ? '…' : ''))
              await handleScanned(value)
            }
          }
        } catch {}
      }
      rafRef.current = requestAnimationFrame(tick)
    }
    log('Detection loop started')
    rafRef.current = requestAnimationFrame(tick)
    return () => { cancelled = true; if (rafRef.current) cancelAnimationFrame(rafRef.current) }
  }, [status?.approved, scanningPaused, drawerOpen])

  // Start/keep camera stream on device change; no need to restart detector
  useEffect(() => {
    if (!status?.approved) return
    ;(async () => {
      try {
        log('Starting camera stream', { deviceId: deviceId || 'default' })
        const constraints: MediaStreamConstraints = deviceId
          ? { video: { deviceId: { exact: deviceId } } as MediaTrackConstraints }
          : { video: true }
        const v = videoRef.current
        if (!v) return
        try { (v.srcObject as MediaStream | null)?.getTracks().forEach((t) => t.stop()) } catch {}
        const stream = await navigator.mediaDevices.getUserMedia(constraints)
        v.srcObject = stream
        await v.play().catch(() => {})
        log('Camera stream active', { tracks: stream.getTracks().map(t => ({ kind: t.kind, state: t.readyState })) })
      } catch {
        setError("Kamerazugriff verweigert oder nicht verfügbar.")
        log('Failed to start camera stream')
      }
    })()
  }, [deviceId, status?.approved])

  async function handleScanned(code: string) {
    setError(null)
    setBusy(true)
    try {
      log('verify-qr start')
      const verify = await fetch(`${API_BASE_URL}/v1/stations/verify-qr`, {
        method: "POST",
        headers: { "Content-Type": "application/json", "X-Station-Key": stationKey },
        body: JSON.stringify({ code }),
      })
      type Problem = { detail?: string }
      if (!verify.ok) {
        const j = (await verify.json().catch(() => ({}))) as Problem
        const msg = j.detail || `Fehler ${verify.status}`
        log('verify-qr failed', msg)
        throw new Error(msg)
      }
      const data = (await verify.json()) as VerifyResult
      setResult(data)
      setScanned(code)
      setDrawerOpen(true)
      log('verify-qr ok', { orderId: data.orderId, items: data.items?.length })
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : "Scan fehlgeschlagen")
      // Verarbeitung wieder zulassen; Kamera läuft weiter
      setScanningPaused(false)
      log('verify-qr error; resume scanning')
    } finally {
      setBusy(false)
    }
  }

  async function redeem() {
    if (!result) return
    const count = result.items?.filter((i) => !i.isRedeemed).length || 0
    if (count === 0) return
    setError(null)
    log('redeem start', { count })
    try {
      const idem = `idem_${Date.now()}_${Math.random().toString(36).slice(2)}`
      const res = await fetch(`${API_BASE_URL}/v1/stations/redeem`, {
        method: "POST",
        headers: { "Content-Type": "application/json", "X-Station-Key": stationKey, "Idempotency-Key": idem },
        body: JSON.stringify({ code: scanned }),
      })
      const json = (await res.json().catch(() => ({}))) as { detail?: string; message?: string }
      if (!res.ok) {
        const msg = json.detail || json.message || `Fehler ${res.status}`
        // Ignore specific backend message when nothing to redeem
        if (/no items to redeem/i.test(msg)) return
        throw new Error(msg)
      }
      log('redeem ok')
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : "Einlösen fehlgeschlagen")
      log('redeem error', e)
    }
  }

  async function requestVerification() {
    const deviceLabel = (devices[0]?.label || navigator.userAgent).slice(0, 80)
    const os = navigator.platform || "web"
    const name = (stationName || deviceLabel || "Station").slice(0, 80)
    if (!name) return
    setBusy(true)
    try {
      log('request verification', { name, os, model: deviceLabel })
      await fetch(`${API_BASE_URL}/v1/stations/requests`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ name, model: deviceLabel, os, deviceKey: stationKey }),
      })
      setRequestSubmitted(true)
      // refresh status
      const r = await fetch(`${API_BASE_URL}/v1/stations/me`, { headers: { "X-Station-Key": stationKey } })
      const js = await r.json()
      setStatus(js as StationStatus)
      log('status after request', js)
    } catch {
      setError("Anfrage fehlgeschlagen")
      log('request verification failed')
    } finally {
      setBusy(false)
    }
  }

  return (
    <div className="bg-background text-foreground flex min-h-screen flex-col items-center gap-4 pt-6 sm:p-6">
      <h1 className="font-primary text-3xl">Abholungsstation</h1>
      <p className="text-muted-foreground max-w-md text-center">Scanne Abhol-QR-Codes hier</p>

      {/* Only show request screen when not approved: nothing else */}
      {status && !status.approved && (
        <div className="border-border bg-card flex w-full max-w-md flex-col gap-4 rounded border p-4">
          {!requestSubmitted ? (
            <>
              <div className="font-semibold">Station-Freigabe anfordern</div>
              <div className="flex gap-2">
                <input
                  className="border-border flex-1 rounded border px-3 py-2"
                  placeholder="Stationsname (z.B. Grill #1)"
                  value={stationName}
                  onChange={(e) => setStationName(e.target.value)}
                />
                <button
                  onClick={requestVerification}
                  className="bg-primary text-primary-foreground rounded px-4 py-2 disabled:opacity-50"
                  disabled={busy || !stationName.trim()}
                >
                  Anfragen
                </button>
              </div>
              {error && (
                <div role="alert" aria-live="assertive" className="text-destructive">
                  {error}
                </div>
              )}
            </>
          ) : (
            <>
              <div className="font-semibold">Anfrage gesendet</div>
              <p className="text-muted-foreground text-sm">
                Deine Stations-Anfrage wurde übermittelt. Sobald sie freigegeben ist, erscheint hier der Kamera-Scanner.
              </p>
            </>
          )}
        </div>
      )}

      {/* After approval: show camera and continuous scanning */}
      {status?.approved && (
        <div className="flex w-full max-w-xl flex-col gap-4">
          {devices.length > 1 && (
            <div className="flex flex-col gap-2">
              <label className="text-sm">Kamera</label>
              <Select value={deviceId} onValueChange={setDeviceId}>
                <SelectTrigger className="w-full">
                  <SelectValue placeholder="Kamera wählen" />
                </SelectTrigger>
                <SelectContent>
                  {devices.map((d) => (
                    <SelectItem key={d.deviceId} value={d.deviceId}>
                      {d.label || `Kamera ${d.deviceId.slice(0, 6)}`}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          )}

          <video ref={videoRef} className="bg-muted w-full rounded-2xl" muted playsInline />
          {error && (
            <div role="alert" aria-live="assertive" className="text-destructive">
              {error}
            </div>
          )}
        </div>
      )}

      {/* Overlay wenn Produkte dargestellt werden: immer Dialog */}
      <Dialog
        open={drawerOpen && !!result}
        onOpenChange={(open) => {
          setDrawerOpen(open)
          if (!open) {
            setResult(null)
            setScanned(null)
            // Kamera bleibt aktiv; wieder Scannen erlauben
            setScanningPaused(false)
            pausedRef.current = false
            lastAtRef.current = 0
            log('Dialog closed; resume scanning')
          }
        }}
      >
        <DialogContent showCloseButton={false}>
          <ModalHeader>
            <ModalTitle>Zu verteilende Artikel</ModalTitle>
          </ModalHeader>
          <div className="mt-2">
            {result && (
              <div className="flex flex-col gap-3">
                {result.items
                  ?.filter((it) => !it.isRedeemed)
                  .map((it) => (
                    <div key={it.id} className="bg-card/50 rounded-xl border p-3">
                      <div className="flex items-center gap-3">
                        {it.productImage ? (
                          <div className="relative h-16 w-16 shrink-0 overflow-hidden rounded-[11px] bg-[#cec9c6]">
                            <Image
                              src={it.productImage}
                              alt={it.title}
                              fill
                              sizes="64px"
                              className="h-full w-full object-cover"
                            />
                          </div>
                        ) : null}
                        <div className="min-w-0 flex-1">
                          <div className="flex items-start justify-between gap-3">
                            <div className="min-w-0">
                              <p className="truncate font-medium">{it.title}</p>
                              {it.menuSlotName && (
                                <div className="text-muted-foreground mt-1 text-xs">
                                  {it.menuSlotName}: {it.title}
                                </div>
                              )}
                            </div>
                            <div className="shrink-0 text-right">
                              <p className="text-muted-foreground text-base">x{it.quantity}</p>
                            </div>
                          </div>
                        </div>
                      </div>
                    </div>
                  ))}
                {(result.items?.filter((it) => !it.isRedeemed).length || 0) === 0 && (
                  <div className="text-muted-foreground text-sm">Keine Artikel zum Einlösen.</div>
                )}
              </div>
            )}
          </div>
          <ModalFooter>
            <Button
              onClick={async () => {
                const count = result?.items?.filter((i) => !i.isRedeemed).length || 0
                if (status?.approved && count > 0) {
                  try {
                    await redeem()
                  } catch {}
                }
                // Ensure scanning resumes even when closing programmatically
                setDrawerOpen(false)
                setResult(null)
                setScanned(null)
                setScanningPaused(false)
                pausedRef.current = false
                lastAtRef.current = 0
                log('Dialog closed; resume scanning')
              }}
            >
              Abschliessen
            </Button>
          </ModalFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}

// Note: continuous scanning is handled via BrowserQRCodeReader.decodeFromVideoDevice
