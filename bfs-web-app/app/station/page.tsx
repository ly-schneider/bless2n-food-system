"use client"

import type { Html5Qrcode } from "html5-qrcode"
import Image from "next/image"
import { useCallback, useEffect, useMemo, useRef, useState } from "react"
import Html5QrcodeScannerPlugin from "@/components/html5-qrcode-scanner-plugin"
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
  const log = useCallback((...args: unknown[]) => console.log("[Station]", ...args), [])
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
  const [cameraPermission, setCameraPermission] = useState<"idle" | "pending" | "granted" | "denied">("idle")
  const [scannerActive, setScannerActive] = useState(false)
  const [scannerKey, setScannerKey] = useState(0)
  const scannerRef = useRef<Html5Qrcode | null>(null)
  const scanLockedRef = useRef(false)

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
    if (!status?.approved) {
      setCameraPermission("idle")
      setScannerActive(false)
      setDevices([])
      setDeviceId("")
    }
  }, [status?.approved])

  const loadDevices = useCallback(async () => {
    try {
      const ds = await navigator.mediaDevices.enumerateDevices()
      const vids = ds.filter((d) => d.kind === "videoinput")
      setDevices(vids)
      log(
        "Cameras",
        vids.map((v) => ({ id: v.deviceId, label: v.label }))
      )
      const preferred = vids[0]?.deviceId || ""
      setDeviceId((prev) => prev || preferred)
      return vids
    } catch (e) {
      setError("Kameras konnten nicht geladen werden.")
      log("enumerateDevices failed", e)
      return []
    }
  }, [log])

  useEffect(() => {
    if (!status?.approved) return
    let cancelled = false
    let permissionStatus: PermissionStatus | null = null

    const applyState = (state: PermissionState) => {
      if (state === "granted") {
        setCameraPermission("granted")
      } else if (state === "denied") {
        setCameraPermission("denied")
      } else {
        setCameraPermission("idle")
      }
    }

    const handleChange = async () => {
      if (!permissionStatus) return
      applyState(permissionStatus.state)
      if (permissionStatus.state === "granted") {
        await loadDevices()
      }
    }

    const checkPermission = async () => {
      try {
        if (!navigator.permissions?.query) return
        permissionStatus = await navigator.permissions.query({ name: "camera" as PermissionName })
        if (cancelled || !permissionStatus) return
        applyState(permissionStatus.state)
        if (permissionStatus.state === "granted") {
          await loadDevices()
        }
        permissionStatus.addEventListener("change", handleChange)
      } catch {
        // ignore; fallback to manual request
      }
    }

    checkPermission()

    return () => {
      cancelled = true
      permissionStatus?.removeEventListener("change", handleChange)
    }
  }, [loadDevices, status?.approved])

  const requestCameraAccess = useCallback(async () => {
    setError(null)
    setScannerActive(false)
    setScannerKey((k) => k + 1)
    setCameraPermission("pending")
    try {
      log("Requesting camera permission")
      const stream = await navigator.mediaDevices.getUserMedia({ video: true })
      await loadDevices()
      setCameraPermission("granted")
      stream.getTracks().forEach((t) => t.stop())
      log("Camera permission granted")
    } catch (e) {
      setCameraPermission("denied")
      setError("Kamerazugriff verweigert oder nicht verfügbar.")
      log("Camera permission failed", e)
    }
  }, [loadDevices, log])

  const handleScannerStartError = useCallback((err: unknown) => {
    setError("Scanner konnte nicht gestartet werden.")
    setScannerActive(false)
    console.error("html5-qrcode start failed", err)
  }, [])

  const handleScannerReady = useCallback(
    (scanner: Html5Qrcode) => {
      scannerRef.current = scanner
      log("html5-qrcode ready")
    },
    [log]
  )

  const handleScannerStop = useCallback(() => {
    scannerRef.current = null
    scanLockedRef.current = false
  }, [])

  const startScanner = useCallback(() => {
    if (!deviceId) {
      setError("Keine Kamera gefunden.")
      return
    }
    setError(null)
    setScannerActive(true)
    setScannerKey((k) => k + 1)
    scanLockedRef.current = false
  }, [deviceId])

  const resumeScanning = useCallback(() => {
    scanLockedRef.current = false
    try {
      scannerRef.current?.resume()
    } catch {}
  }, [])

  const handleScanned = useCallback(
    async (code: string) => {
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
        resumeScanning()
        log("verify-qr error; resume scanning", e)
      }
    },
    [resumeScanning, stationKey, log]
  )

  const handleDecoded = useCallback(
    async (decodedText: string) => {
      if (!decodedText) return
      if (scanLockedRef.current) return
      scanLockedRef.current = true
      try {
        scannerRef.current?.pause(true)
      } catch {}
      await handleScanned(decodedText)
    },
    [handleScanned]
  )

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
      <p className="text-muted-foreground mb-4 max-w-md text-center">Scanne Abhol-QR-Codes hier</p>

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
        <div className="flex w-full flex-col gap-4 sm:max-w-xl">
          {!scannerActive && (
            <div className="bg-background rounded-xl border p-5 shadow-sm">
              <div className="flex flex-col gap-2">
                <h2 className="text-xl font-semibold">Scanner starten</h2>
              </div>

              <div className="mt-2 grid gap-3">
                <div className="flex flex-wrap gap-2">
                  {cameraPermission !== "granted" && (
                    <Button onClick={requestCameraAccess} disabled={cameraPermission === "pending"}>
                      {cameraPermission === "pending" ? "Kamerazugriff wird angefragt..." : "Kamerazugriff erlauben"}
                    </Button>
                  )}
                </div>

                {cameraPermission === "granted" && (
                  <div className="grid gap-2">
                    <Label htmlFor="camera">Kamera</Label>
                    {devices.length > 0 ? (
                      <>
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
                        <Button
                          variant="secondary"
                          onClick={startScanner}
                          disabled={cameraPermission !== "granted" || !deviceId}
                        >
                          Scanner starten
                        </Button>
                      </>
                    ) : (
                      <p className="text-muted-foreground text-sm">Keine Kamera gefunden.</p>
                    )}
                  </div>
                )}

                {cameraPermission === "denied" && (
                  <div className="rounded-xl border border-amber-200 bg-amber-50 p-3 text-sm text-amber-700">
                    Zugriff abgelehnt. Bitte erlaube den Kamerazugriff in den Browser-Einstellungen und versuche es
                    erneut.
                  </div>
                )}
              </div>
            </div>
          )}

          {scannerActive && cameraPermission === "granted" && (
            <div className="w-full">
              <Html5QrcodeScannerPlugin
                key={`${deviceId}-${scannerKey}`}
                cameraId={deviceId || undefined}
                fps={10}
                qrbox={250}
                disableFlip={false}
                qrCodeSuccessCallback={handleDecoded}
                onReady={handleScannerReady}
                onStartError={handleScannerStartError}
                onStop={handleScannerStop}
              />
            </div>
          )}

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
            resumeScanning()
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
                resumeScanning()
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
