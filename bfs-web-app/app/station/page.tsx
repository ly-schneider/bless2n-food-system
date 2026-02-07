"use client"

import type { Html5Qrcode } from "html5-qrcode"
import Image from "next/image"
import { useCallback, useEffect, useRef, useState } from "react"
import { PairingCodeDisplay } from "@/components/device/pairing-code-display"
import Html5QrcodeScannerPlugin from "@/components/html5-qrcode-scanner-plugin"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogFooter as ModalFooter,
  DialogHeader as ModalHeader,
  DialogTitle as ModalTitle,
} from "@/components/ui/dialog"
import { Label } from "@/components/ui/label"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"

import { getDeviceToken } from "@/lib/device-auth"

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
  const [drawerOpen, setDrawerOpen] = useState(false)
  const [cameraPermission, setCameraPermission] = useState<"idle" | "pending" | "granted" | "denied">("idle")
  const [scannerActive, setScannerActive] = useState(false)
  const [scannerKey, setScannerKey] = useState(0)
  const scannerRef = useRef<Html5Qrcode | null>(null)
  const scanLockedRef = useRef(false)

  // sessionToken = the authenticated Bearer token from device-auth
  const [sessionToken] = useState<string | null>(() => getDeviceToken())

  const bearerToken = sessionToken || ""

  useEffect(() => {
    if (!sessionToken) {
      setStatus({ exists: false, approved: false })
      log("no session token; showing enrollment")
      return
    }
    ;(async () => {
      try {
        log("GET /v1/stations/me")
        const res = await fetch(`/api/v1/stations/me`, {
          headers: { Authorization: `Bearer ${sessionToken}` },
        })
        if (res.status === 401 || res.status === 403) {
          setStatus({ exists: false, approved: false })
          log("status auth failed; showing enrollment")
          return
        }
        const json = await res.json()
        setStatus(json as StationStatus)
        log("status", json)
      } catch {
        setStatus({ exists: false, approved: false })
        log("status fetch failed")
      }
    })()
  }, [sessionToken])

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

  // Remove old enrollment / polling logic — replaced by PairingCodeDisplay

  const handleScanned = useCallback(
    async (code: string) => {
      setError(null)
      try {
        // Client-side: extract orderId from scanned QR (raw orderId string)
        const orderId = code.trim()
        if (!orderId) throw new Error("Ungültiger QR-Code")
        log("order lookup start", { orderId })

        // Fetch order details to display items
        const orderRes = await fetch(`/api/v1/orders/${encodeURIComponent(orderId)}`, {
          headers: { Authorization: `Bearer ${bearerToken}` },
        })
        type Problem = { detail?: string }
        if (!orderRes.ok) {
          const j = (await orderRes.json().catch(() => ({}))) as Problem
          const msg = j.detail || `Fehler ${orderRes.status}`
          log("order lookup failed", msg)
          throw new Error(msg)
        }
        const orderData = (await orderRes.json()) as { order: { id: string; items: PublicOrderItem[] } }
        const data: VerifyResult = { orderId: orderData.order.id, items: orderData.order.items }
        setResult(data)
        setScanned(orderId)
        setDrawerOpen(true)
        log("order lookup ok", { orderId: data.orderId, items: data.items?.length })
      } catch (e: unknown) {
        setError(e instanceof Error ? e.message : "Scan fehlgeschlagen")
        resumeScanning()
        log("order lookup error; resume scanning", e)
      }
    },
    [resumeScanning, bearerToken, log]
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
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${bearerToken}`,
          "Idempotency-Key": idem,
        },
        body: JSON.stringify({ orderId: scanned }),
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

  return (
    <div className="bg-background text-foreground flex min-h-screen flex-col items-center gap-2 pt-6 sm:p-6">
      <h1 className="font-primary text-3xl">Abholungsstation</h1>
      <p className="text-muted-foreground mb-4 max-w-md text-center">Scanne Abhol-QR-Codes hier</p>

      {status && !status.approved && <PairingCodeDisplay deviceType="STATION" />}

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
