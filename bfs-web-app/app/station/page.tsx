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

const UUID_RE = /[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}/i
const CAMPAIGN_PREFIX = "CAMP:"

type ParsedScan = { kind: "order" | "campaign"; id: string }

function parseScan(raw: string): ParsedScan | null {
  const trimmed = (raw ?? "").trim()
  if (!trimmed) return null
  if (trimmed.toUpperCase().startsWith(CAMPAIGN_PREFIX)) {
    const match = trimmed.slice(CAMPAIGN_PREFIX.length).match(UUID_RE)
    return match ? { kind: "campaign", id: match[0] } : null
  }
  const match = trimmed.match(UUID_RE)
  return match ? { kind: "order", id: match[0] } : null
}

type StationProduct = { productId: string; name?: string }
type StationStatus = {
  id: string
  status: "pending" | "approved" | "rejected" | "revoked"
  name?: string
  products?: StationProduct[]
}

export default function StationPage() {
  const [devices, setDevices] = useState<MediaDeviceInfo[]>([])
  const [deviceId, setDeviceId] = useState<string>("")
  const [status, setStatus] = useState<StationStatus | null>(null)
  const [systemDisabled, setSystemDisabled] = useState(false)
  const [error, setError] = useState<string | null>(null)
  type OrderLine = {
    id: string
    orderId: string
    productId: string
    lineType?: string
    title: string
    quantity: number
    unitPriceCents: number
    productImage?: string | null
    redemption?: { id: string; redeemedAt: string } | null
    parentLineId?: string | null
    menuSlotId?: string | null
    menuSlotName?: string | null
    childLines?: OrderLine[]
  }
  type VerifyResult = { orderId: string; lines: OrderLine[] }
  const [result, setResult] = useState<VerifyResult | null>(null)
  const [scanKind, setScanKind] = useState<"order" | "campaign" | null>(null)
  const [scanned, setScanned] = useState<string | null>(null)
  const [drawerOpen, setDrawerOpen] = useState(false)
  const [dialogLoading, setDialogLoading] = useState(false)
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
      setStatus(null)
      return
    }
    ;(async () => {
      try {
        const res = await fetch(`/api/v1/stations/me`, {
          headers: { Authorization: `Bearer ${sessionToken}` },
        })
        if (res.status === 503) {
          setSystemDisabled(true)
          return
        }
        if (res.status === 401 || res.status === 403) {
          setStatus(null)
          return
        }
        setSystemDisabled(false)
        const json = await res.json()
        setStatus(json as StationStatus)
      } catch {
        setStatus(null)
      }
    })()
  }, [sessionToken])

  useEffect(() => {
    if (status?.status !== "approved") {
      setCameraPermission("idle")
      setScannerActive(false)
      setDevices([])
      setDeviceId("")
    }
  }, [status?.status === "approved"])

  const loadDevices = useCallback(async () => {
    try {
      const ds = await navigator.mediaDevices.enumerateDevices()
      const vids = ds.filter((d) => d.kind === "videoinput")
      setDevices(vids)
      const preferred = vids[0]?.deviceId || ""
      setDeviceId((prev) => prev || preferred)
      return vids
    } catch {
      setError("Kameras konnten nicht geladen werden.")
      return []
    }
  }, [])

  useEffect(() => {
    if (status?.status !== "approved") return
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
  }, [loadDevices, status?.status === "approved"])

  const requestCameraAccess = useCallback(async () => {
    setError(null)
    setScannerActive(false)
    setScannerKey((k) => k + 1)
    setCameraPermission("pending")
    try {
      const stream = await navigator.mediaDevices.getUserMedia({ video: true })
      await loadDevices()
      setCameraPermission("granted")
      stream.getTracks().forEach((t) => t.stop())
    } catch {
      setCameraPermission("denied")
      setError("Kamerazugriff verweigert oder nicht verfügbar.")
    }
  }, [loadDevices])

  const handleScannerStartError = useCallback(() => {
    setError("Scanner konnte nicht gestartet werden.")
    setScannerActive(false)
  }, [])

  const handleScannerReady = useCallback((scanner: Html5Qrcode) => {
    scannerRef.current = scanner
  }, [])

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
      const parsed = parseScan(code)
      if (!parsed) {
        setError("Ungültiger QR-Code")
        resumeScanning()
        return
      }

      setScanned(parsed.id)
      setScanKind(parsed.kind)
      setResult(null)
      setDialogLoading(true)
      setDrawerOpen(true)

      try {
        if (parsed.kind === "campaign") {
          const idem = `idem_${Date.now()}_${Math.random().toString(36).slice(2)}`
          const res = await fetch(`/api/v1/stations/redeem-campaign`, {
            method: "POST",
            headers: {
              "Content-Type": "application/json",
              Authorization: `Bearer ${bearerToken}`,
              "Idempotency-Key": idem,
            },
            body: JSON.stringify({ claimToken: parsed.id }),
          })
          type Problem = { detail?: string; message?: string }
          if (!res.ok) {
            const j = (await res.json().catch(() => ({}))) as Problem
            if (res.status === 409) {
              throw new Error(j.detail || j.message || "Maximale Einlösungen erreicht.")
            }
            if (res.status === 410) {
              throw new Error(j.detail || j.message || "Kampagne nicht mehr aktiv.")
            }
            if (res.status === 404) {
              throw new Error(j.detail || j.message || "Kampagne nicht gefunden.")
            }
            throw new Error(j.detail || j.message || `Fehler ${res.status}`)
          }
          type CampaignResp = {
            orderId: string
            redemptionCount: number
            maxRedemptions: number
            station?: {
              orderId?: string
              redeemedAt?: string
              items?: Array<{
                id: string
                orderId: string
                productId: string
                title: string
                quantity: number
                isRedeemed: boolean
                parentItemId?: string | null
                menuSlotId?: string | null
                menuSlotName?: string | null
              }>
            }
          }
          const data = (await res.json()) as CampaignResp
          const redeemedAt = data.station?.redeemedAt || new Date().toISOString()
          const items = data.station?.items || []
          const lines: OrderLine[] = items.map((it) => ({
            id: it.id,
            orderId: it.orderId,
            productId: it.productId,
            title: it.title,
            quantity: it.quantity,
            unitPriceCents: 0,
            redemption: it.isRedeemed ? { id: it.id, redeemedAt } : null,
            parentLineId: it.parentItemId ?? null,
            menuSlotId: it.menuSlotId ?? null,
            menuSlotName: it.menuSlotName ?? null,
          }))
          setResult({ orderId: data.orderId, lines })
          return
        }

        const orderRes = await fetch(`/api/v1/orders/${encodeURIComponent(parsed.id)}`, {
          headers: { Authorization: `Bearer ${bearerToken}` },
        })
        type Problem = { detail?: string }
        if (!orderRes.ok) {
          const j = (await orderRes.json().catch(() => ({}))) as Problem
          throw new Error(j.detail || `Fehler ${orderRes.status}`)
        }
        const orderData = (await orderRes.json()) as { id: string; lines?: OrderLine[] }
        const allLines = orderData.lines || []
        const assignedProductIds = new Set((status?.products || []).map((p) => p.productId))
        const matchedBundleIds = new Set(
          allLines.filter((l) => assignedProductIds.has(l.productId) && l.lineType === "bundle").map((l) => l.id)
        )
        const stationLines = allLines.filter(
          (l) =>
            (assignedProductIds.has(l.productId) && l.lineType !== "bundle") ||
            (l.parentLineId != null && matchedBundleIds.has(l.parentLineId))
        )
        setResult({ orderId: orderData.id, lines: stationLines })
      } catch (e: unknown) {
        setError(e instanceof Error ? e.message : "Scan fehlgeschlagen")
        setDrawerOpen(false)
        resumeScanning()
      } finally {
        setDialogLoading(false)
      }
    },
    [resumeScanning, bearerToken, status]
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

  const fireRedeem = useCallback(
    (orderId: string) => {
      const idem = `idem_${Date.now()}_${Math.random().toString(36).slice(2)}`
      fetch(`/api/v1/stations/redeem`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${bearerToken}`,
          "Idempotency-Key": idem,
        },
        body: JSON.stringify({ orderId }),
      }).catch(() => {})
    },
    [bearerToken]
  )

  useEffect(() => {
    if (!systemDisabled) return
    const interval = setInterval(async () => {
      try {
        const res = await fetch(`/api/v1/system/status`)
        if (!res.ok) return
        const data = (await res.json()) as { enabled: boolean }
        if (data.enabled) {
          setSystemDisabled(false)
          window.location.reload()
        }
      } catch {}
    }, 30_000)
    return () => clearInterval(interval)
  }, [systemDisabled])

  if (systemDisabled) {
    return (
      <div className="bg-background flex min-h-screen items-center justify-center">
        <div className="text-center">
          <h2 className="text-2xl font-semibold">System geschlossen</h2>
          <p className="text-muted-foreground mt-2">Das System ist momentan nicht verfügbar.</p>
        </div>
      </div>
    )
  }

  return (
    <div className="bg-background text-foreground flex flex-1 flex-col items-center gap-2 pt-6 sm:p-6">
      <h1 className="font-primary text-3xl">Abholungsstation</h1>
      {status?.name && <p className="text-lg font-medium">{status.name}</p>}
      <p className="text-muted-foreground max-w-md text-center">Scanne Abhol-QR-Codes hier</p>

      {(!status || status.status !== "approved") && (
        <div className="flex w-full flex-1 items-center justify-center">
          <PairingCodeDisplay deviceType="STATION" />
        </div>
      )}

      {status?.status === "approved" && (
        <div className="mt-8 flex w-full flex-col gap-4 sm:max-w-xl">
          {!scannerActive && (
            <div className="bg-background rounded-xl border p-5 shadow-sm">
              <h2 className="text-center text-xl font-semibold">Scanner starten</h2>

              <div className="mt-3 grid justify-items-center gap-3">
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
        open={drawerOpen}
        onOpenChange={(open) => {
          if (!open) {
            setDrawerOpen(false)
            setResult(null)
            setScanned(null)
            setScanKind(null)
            setDialogLoading(false)
            resumeScanning()
          }
        }}
      >
        <DialogContent showCloseButton={false}>
          <ModalHeader>
            <ModalTitle>Zu verteilende Artikel</ModalTitle>
          </ModalHeader>
          <div className="mt-2">
            {dialogLoading && (
              <div className="flex flex-col gap-3">
                {[0, 1].map((i) => (
                  <div key={i} className="rounded-xl border p-3">
                    <div className="flex items-center gap-3">
                      <div className="h-16 w-16 shrink-0 animate-pulse rounded-[11px] bg-gray-200" />
                      <div className="flex-1 space-y-2">
                        <div className="h-4 w-3/4 animate-pulse rounded bg-gray-200" />
                        <div className="h-3 w-1/2 animate-pulse rounded bg-gray-200" />
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            )}
            {!dialogLoading &&
              result &&
              (() => {
                const matched = result.lines || []
                const unredeemed = matched.filter((it) => !it.redemption)
                const alreadyRedeemed = matched.filter((it) => it.redemption)
                const renderLineCard = (it: OrderLine, redeemed: boolean) => (
                  <div
                    key={it.id}
                    className={`rounded-xl border p-3 ${redeemed ? "bg-muted/40 opacity-70" : "bg-card/50"}`}
                  >
                    <div className="flex items-center gap-3">
                      {it.productImage ? (
                        <div className="relative h-16 w-16 shrink-0 overflow-hidden rounded-[11px] bg-[#cec9c6]">
                          <Image
                            src={it.productImage}
                            alt={it.title}
                            fill
                            sizes="64px"
                            className={`h-full w-full object-cover ${redeemed ? "grayscale" : ""}`}
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
                            {redeemed && it.redemption?.redeemedAt && (
                              <div className="text-muted-foreground mt-1 text-xs">
                                Eingelöst: {new Date(it.redemption.redeemedAt).toLocaleString("de-CH")}
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
                )

                if (matched.length === 0) {
                  return (
                    <div className="bg-muted/40 text-muted-foreground rounded-xl border p-4 text-sm">
                      Diese Station bietet keine Artikel aus dieser Bestellung an.
                    </div>
                  )
                }

                if (unredeemed.length === 0) {
                  return (
                    <div className="flex flex-col gap-3">
                      <div className="rounded-xl border border-amber-300 bg-amber-50 p-3 text-sm font-medium text-amber-900">
                        Bereits eingelöst
                      </div>
                      {alreadyRedeemed.map((it) => renderLineCard(it, true))}
                    </div>
                  )
                }

                return (
                  <div className="flex flex-col gap-3">
                    {unredeemed.map((it) => renderLineCard(it, false))}
                    {alreadyRedeemed.length > 0 && (
                      <>
                        <div className="text-muted-foreground mt-2 text-xs tracking-wide uppercase">
                          Bereits eingelöst
                        </div>
                        {alreadyRedeemed.map((it) => renderLineCard(it, true))}
                      </>
                    )}
                  </div>
                )
              })()}
          </div>
          <ModalFooter>
            <Button
              disabled={dialogLoading}
              onClick={() => {
                const currentOrderId = result?.orderId || scanned
                const hasItems = (result?.lines?.filter((i) => !i.redemption).length || 0) > 0
                const kind = scanKind

                setDrawerOpen(false)
                setResult(null)
                setScanned(null)
                setScanKind(null)
                resumeScanning()

                if (kind === "order" && status?.status === "approved" && hasItems && currentOrderId) {
                  fireRedeem(currentOrderId)
                }
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
