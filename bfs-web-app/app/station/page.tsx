"use client"

import { BarcodeFormat, BrowserMultiFormatReader } from "@zxing/browser"
import type { IScannerControls } from "@zxing/browser"
import { DecodeHintType } from "@zxing/library"
import Image from "next/image"
import { useEffect, useMemo, useRef, useState } from "react"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogFooter as ModalFooter,
  DialogHeader as ModalHeader,
  DialogTitle as ModalTitle,
} from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"

import { getClientInfo } from "@/lib/client-info"
import { randomUrlSafe } from "@/lib/crypto"

type StationStatus = { exists: boolean; approved: boolean; name?: string }

export default function StationPage() {
  const log = (...args: unknown[]) => console.log("[Station]", ...args)
  const videoRef = useRef<HTMLVideoElement | null>(null)
  const [devices, setDevices] = useState<MediaDeviceInfo[]>([])
  const [deviceId, setDeviceId] = useState<string>("")
  const [status, setStatus] = useState<StationStatus | null>(null)
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
  // Prefer native BarcodeDetector with ZXing fallback for older browsers
  const [scanningPaused, setScanningPaused] = useState(false)
  const pausedRef = useRef(false)
  interface DetectedBarcode {
    rawValue?: string
  }
  interface BarcodeDetector {
    detect: (
      source: HTMLVideoElement | CanvasImageSource | ImageBitmap | ImageData | Blob
    ) => Promise<DetectedBarcode[]>
  }
  type BarcodeDetectorConstructor = new (init?: { formats?: string[] }) => BarcodeDetector
  const detectorRef = useRef<BarcodeDetector | null>(null)
  const rafRef = useRef<number | null>(null)
  const lastCodeRef = useRef<string | null>(null)
  const lastAtRef = useRef<number>(0)
  const zxingRef = useRef<BrowserMultiFormatReader | null>(null)
  const zxingControlsRef = useRef<IScannerControls | null>(null)
  const [hasNativeDetector, setHasNativeDetector] = useState<boolean | null>(null)
  useEffect(() => {
    pausedRef.current = scanningPaused || drawerOpen
  }, [scanningPaused, drawerOpen])

  const stationKey = useMemo(() => {
    if (typeof window === "undefined") return ""
    let k = localStorage.getItem("bfs.stationKey")
    if (!k) {
      k = `st_${randomUrlSafe(24)}`
      localStorage.setItem("bfs.stationKey", k)
    }
    log("stationKey", k)
    return k
  }, [])
  // No offline mode or service worker – simplify flow

  useEffect(() => {
    // fetch station status
    ;(async () => {
      try {
        log("GET /v1/stations/me")
        const res = await fetch(`/api/v1/stations/me`, { headers: { "X-Station-Key": stationKey } })
        const json = await res.json()
        setStatus(json as StationStatus)
        log("status", json)
      } catch {
        setStatus({ exists: false, approved: false })
        log("status fetch failed")
      }
    })()
  }, [stationKey])

  useEffect(() => {
    if (typeof window === "undefined") return
    const BD = (window as unknown as { BarcodeDetector?: BarcodeDetectorConstructor }).BarcodeDetector ?? undefined
    const supported = !!BD
    setHasNativeDetector(supported)
    if (!supported) log("Native BarcodeDetector missing; will use ZXing fallback")
  }, [])

  // After approval: request camera permission, enumerate devices, and start scanning
  useEffect(() => {
    if (!status?.approved) return
    ;(async () => {
      try {
        // Request permission with default camera to reveal labels
        log("Requesting camera permission")
        await navigator.mediaDevices.getUserMedia({ video: true })
        log("Camera permission granted")
        const ds = await navigator.mediaDevices.enumerateDevices()
        const vids = ds.filter((d) => d.kind === "videoinput")
        setDevices(vids)
        log(
          "Cameras",
          vids.map((v) => ({ id: v.deviceId, label: v.label }))
        )
        const preferred = vids[0]?.deviceId || ""
        setDeviceId((prev) => prev || preferred)
      } catch (e) {
        setError("Kamerazugriff verweigert oder nicht verfügbar.")
        log("Camera permission failed", e)
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
    const BD =
      typeof window !== "undefined"
        ? (window as unknown as { BarcodeDetector?: BarcodeDetectorConstructor }).BarcodeDetector
        : undefined
    if (!BD || hasNativeDetector !== true) return
    if (!detectorRef.current) {
      try {
        detectorRef.current = new BD({ formats: ["qr_code"] })
        log("BarcodeDetector initialized")
      } catch {
        return
      }
    }
    let cancelled = false
    const tick = async () => {
      if (cancelled) return
      if (!(scanningPaused || drawerOpen) && videoRef.current && detectorRef.current) {
        try {
          const codes = await detectorRef.current.detect(videoRef.current)
          const found = codes?.find((c) => c.rawValue)
          if (found?.rawValue) {
            const now = Date.now()
            const value = String(found.rawValue)
            const isRepeat = lastCodeRef.current === value && now - lastAtRef.current < 1500
            if (!isRepeat) {
              lastCodeRef.current = value
              lastAtRef.current = now
              setScanningPaused(true)
              log("QR detected", value.slice(0, 24) + (value.length > 24 ? "…" : ""))
              await handleScanned(value)
            }
          }
        } catch {}
      }
      rafRef.current = requestAnimationFrame(tick)
    }
    log("Detection loop started")
    rafRef.current = requestAnimationFrame(tick)
    return () => {
      cancelled = true
      if (rafRef.current) cancelAnimationFrame(rafRef.current)
    }
  }, [status?.approved, scanningPaused, drawerOpen, hasNativeDetector])

  // ZXing fallback for browsers without BarcodeDetector (e.g., older Safari)
  useEffect(() => {
    if (!status?.approved) return
    if (hasNativeDetector !== false) return
    const v = videoRef.current
    if (!v) return
    if (!zxingRef.current) {
      const hints = new Map<DecodeHintType, unknown>()
      hints.set(DecodeHintType.POSSIBLE_FORMATS, [BarcodeFormat.QR_CODE])
      zxingRef.current = new BrowserMultiFormatReader(hints, { delayBetweenScanAttempts: 250 })
      log("ZXing reader initialized")
    }
    const reader = zxingRef.current
    let cancelled = false
    reader
      .decodeFromVideoElement(v, async (result, err, controls) => {
        if (cancelled) return
        if (!zxingControlsRef.current && controls) zxingControlsRef.current = controls
        if (result && !(scanningPaused || drawerOpen)) {
          const value = result.getText()
          const now = Date.now()
          const isRepeat = lastCodeRef.current === value && now - lastAtRef.current < 1500
          if (!isRepeat && value) {
            lastCodeRef.current = value
            lastAtRef.current = now
            setScanningPaused(true)
            log("QR detected (fallback)", value.slice(0, 24) + (value.length > 24 ? "…" : ""))
            await handleScanned(value)
          }
        } else if (err && (err as Error).name !== "NotFoundException") {
          log("ZXing error", err)
        }
      })
      .catch(() => log("ZXing decode failed to start"))
    return () => {
      cancelled = true
      zxingControlsRef.current?.stop?.()
      zxingControlsRef.current = null
      ;(reader as { reset?: () => void }).reset?.()
    }
  }, [status?.approved, hasNativeDetector, deviceId])

  // Start/keep camera stream on device change; no need to restart detector
  useEffect(() => {
    if (!status?.approved) return
    ;(async () => {
      try {
        log("Starting camera stream", { deviceId: deviceId || "default" })
        const baseVideo: MediaTrackConstraints = {
          facingMode: { ideal: "environment" },
          // focusMode is non-standard but helps iOS Safari auto-focus when supported
          focusMode: "continuous",
        } as MediaTrackConstraints
        const constraints: MediaStreamConstraints = deviceId
          ? { video: { ...baseVideo, deviceId: { exact: deviceId } } as MediaTrackConstraints }
          : { video: baseVideo }
        const v = videoRef.current
        if (!v) return
        try {
          ;(v.srcObject as MediaStream | null)?.getTracks().forEach((t) => t.stop())
        } catch {}
        const stream = await navigator.mediaDevices.getUserMedia(constraints)
        v.srcObject = stream
        await v.play().catch(() => {})
        const track = stream.getVideoTracks()[0]
        const capabilities = track?.getCapabilities?.() as MediaTrackCapabilities & { focusMode?: string[] }
        const focusPrefs: MediaTrackConstraints[] = []
        if (capabilities?.focusMode?.includes("continuous")) {
          focusPrefs.push({ focusMode: "continuous" } as MediaTrackConstraints)
        } else if (capabilities?.focusMode?.includes("auto")) {
          focusPrefs.push({ focusMode: "auto" } as MediaTrackConstraints)
        }
        if (focusPrefs.length && track?.applyConstraints) {
          await track.applyConstraints({ advanced: focusPrefs }).catch(() => {})
        }
        log("Camera stream active", {
          tracks: stream.getTracks().map((t) => ({ kind: t.kind, state: t.readyState })),
          focus: capabilities?.focusMode,
        })
      } catch {
        setError("Kamerazugriff verweigert oder nicht verfügbar.")
        log("Failed to start camera stream")
      }
    })()
  }, [deviceId, status?.approved])

  async function handleScanned(code: string) {
    setError(null)
    try {
      log("verify-qr start")
      const verify = await fetch(`/api/v1/stations/verify-qr`, {
        method: "POST",
        headers: { "Content-Type": "application/json", "X-Station-Key": stationKey },
        body: JSON.stringify({ code }),
      })
      type Problem = { detail?: string }
      if (!verify.ok) {
        const j = (await verify.json().catch(() => ({}))) as Problem
        const msg = j.detail || `Fehler ${verify.status}`
        log("verify-qr failed", msg)
        throw new Error(msg)
      }
      const data = (await verify.json()) as VerifyResult
      setResult(data)
      setScanned(code)
      setDrawerOpen(true)
      log("verify-qr ok", { orderId: data.orderId, items: data.items?.length })
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : "Scan fehlgeschlagen")
      // Verarbeitung wieder zulassen; Kamera läuft weiter
      setScanningPaused(false)
      log("verify-qr error; resume scanning")
    }
  }

  async function redeem() {
    if (!result) return
    const count = result.items?.filter((i) => !i.isRedeemed).length || 0
    if (count === 0) return
    setError(null)
    log("redeem start", { count })
    try {
      const idem = `idem_${Date.now()}_${Math.random().toString(36).slice(2)}`
      const res = await fetch(`/api/v1/stations/redeem`, {
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
      log("redeem ok")
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : "Einlösen fehlgeschlagen")
      log("redeem error", e)
    }
  }

  async function requestVerification() {
    const fallbackLabel = devices[0]?.label || ""
    const info = await getClientInfo()
    const os = info.os || "web"
    const name = (stationName || fallbackLabel || "Station").slice(0, 80)
    if (!name) return
    try {
      log("request verification", { name, os, model: info.model })
      await fetch(`/api/v1/stations/requests`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ name, model: info.model, os, deviceKey: stationKey }),
      })
      setRequestSubmitted(true)
      // refresh status
      const r = await fetch(`/api/v1/stations/me`, { headers: { "X-Station-Key": stationKey } })
      const js = await r.json()
      setStatus(js as StationStatus)
      log("status after request", js)
    } catch {
      setError("Anfrage fehlgeschlagen")
      log("request verification failed")
    }
  }

  return (
    <div className="bg-background text-foreground flex min-h-screen flex-col items-center gap-2 pt-6 sm:p-6">
      <h1 className="font-primary text-3xl">Abholungsstation</h1>
      <p className="text-muted-foreground max-w-md text-center">Scanne Abhol-QR-Codes hier</p>

      {status && !status.approved && (
        <div className="place-items-center p-4">
          <div className="bg-background w-full max-w-md rounded-xl border p-5 shadow-sm">
            <h1 className="mb-2 text-2xl font-semibold">Gerät registrieren</h1>
            <p className="text-muted-foreground mb-4 text-sm">
              Dieses Gerät muss vor dem Verkauf von einem Admin freigegeben werden.
            </p>
            <div className="grid gap-3">
              <div className="grid gap-1">
                <Label htmlFor="name">Gerätename</Label>
                <Input
                  id="name"
                  value={stationName}
                  onChange={(e) => setStationName(e.target.value)}
                  placeholder="z. B. Grill 1"
                />
              </div>
              <Button onClick={requestVerification}>Zugang anfordern</Button>
              {requestSubmitted && (
                <div className="rounded-xl border border-amber-200 bg-amber-50 p-2 text-sm text-amber-700">
                  Anfrage gesendet. Ein Admin muss dieses Gerät freigeben.
                </div>
              )}
            </div>
          </div>
        </div>
      )}

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

      <Dialog
        open={drawerOpen && !!result}
        onOpenChange={(open) => {
          setDrawerOpen(open)
          if (!open) {
            setResult(null)
            setScanned(null)
            setScanningPaused(false)
            pausedRef.current = false
            lastAtRef.current = 0
            log("Dialog closed; resume scanning")
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
                log("Dialog closed; resume scanning")
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
