"use client"

import type { Html5Qrcode } from "html5-qrcode"
import { AlertCircle, CheckCircle2, Info, Loader2, PackageOpen, ScanLine, XCircle } from "lucide-react"
import Image from "next/image"
import { useCallback, useEffect, useRef, useState } from "react"
import { PairingCodeDisplay } from "@/components/device/pairing-code-display"
import Html5QrcodeScannerPlugin from "@/components/html5-qrcode-scanner-plugin"
import { FullscreenButton } from "@/components/station/fullscreen-button"
import { HardRefreshButton } from "@/components/station/hard-refresh-button"
import { Button } from "@/components/ui/button"
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import { Label } from "@/components/ui/label"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"

import { getDeviceToken } from "@/lib/device-auth"
import { playScanSound, primeScanAudio } from "@/lib/scan-sound"

const UUID_RE = /[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}/i
const CAMPAIGN_PREFIX = "CAMP:"
const SCAN_RESUME_DELAY_MS = 1400
const DUPLICATE_SCAN_WINDOW_MS = 5000
const HISTORY_LIMIT = 8

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

type ScanStatus = "loading" | "success" | "already-redeemed" | "no-items" | "error"

type ScanResult = {
  key: string
  scanId: string
  kind: "order" | "campaign"
  status: ScanStatus
  message?: string
  scannedAt: Date
  orderId?: string
  lines: OrderLine[]
  bundleParents?: Record<string, OrderLine>
}

type LineGroup = { key: string; parent: OrderLine; children: OrderLine[] }

function groupLines(lines: OrderLine[], bundleParents?: Record<string, OrderLine>): LineGroup[] {
  const groups: LineGroup[] = []
  const groupByRootId: Record<string, LineGroup> = {}
  for (const l of lines) {
    const parentLine = l.parentLineId && bundleParents ? bundleParents[l.parentLineId] : null
    if (parentLine) {
      let g = groupByRootId[parentLine.id]
      if (!g) {
        g = { key: parentLine.id, parent: parentLine, children: [] }
        groupByRootId[parentLine.id] = g
        groups.push(g)
      }
      g.children.push(l)
    } else {
      groups.push({ key: l.id, parent: l, children: [] })
    }
  }
  return groups
}

function groupRedemptionTime(group: LineGroup): string | undefined {
  if (group.parent.redemption?.redeemedAt) return group.parent.redemption.redeemedAt
  return group.children.find((c) => c.redemption?.redeemedAt)?.redemption?.redeemedAt
}

type ConsolidatedItem = {
  productId: string
  title: string
  productImage?: string | null
  quantity: number
  redeemed: boolean
  redemptionTime?: string
}

function consolidateLines(lines: OrderLine[]): ConsolidatedItem[] {
  const map = new Map<string, ConsolidatedItem>()
  for (const l of lines) {
    const existing = map.get(l.productId)
    if (existing) {
      existing.quantity += l.quantity
      if (!l.redemption) existing.redeemed = false
      const t = l.redemption?.redeemedAt
      if (t && (!existing.redemptionTime || t > existing.redemptionTime)) {
        existing.redemptionTime = t
      }
    } else {
      map.set(l.productId, {
        productId: l.productId,
        title: l.title,
        productImage: l.productImage,
        quantity: l.quantity,
        redeemed: !!l.redemption,
        redemptionTime: l.redemption?.redeemedAt,
      })
    }
  }
  return Array.from(map.values())
}

type StatusMeta = {
  label: string
  text: string
  bg: string
  dot: string
  icon: typeof CheckCircle2
}

const STATUS_STYLES: Record<ScanStatus, StatusMeta> = {
  loading: {
    label: "Wird geprüft",
    text: "text-muted-foreground",
    bg: "bg-card",
    dot: "bg-muted-foreground/50",
    icon: Loader2,
  },
  success: {
    label: "Eingelöst",
    text: "text-success",
    bg: "bg-success/10",
    dot: "bg-success",
    icon: CheckCircle2,
  },
  "already-redeemed": {
    label: "Bereits eingelöst",
    text: "text-destructive",
    bg: "bg-destructive/10",
    dot: "bg-destructive",
    icon: AlertCircle,
  },
  "no-items": {
    label: "Keine Artikel",
    text: "text-muted-foreground",
    bg: "bg-card",
    dot: "bg-muted-foreground/50",
    icon: PackageOpen,
  },
  error: {
    label: "Fehler",
    text: "text-destructive",
    bg: "bg-destructive/10",
    dot: "bg-destructive",
    icon: XCircle,
  },
}

export default function StationPage() {
  const [devices, setDevices] = useState<MediaDeviceInfo[]>([])
  const [deviceId, setDeviceId] = useState<string>("")
  const [status, setStatus] = useState<StationStatus | null>(null)
  const [systemDisabled, setSystemDisabled] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [scans, setScans] = useState<ScanResult[]>([])
  const [detailOrderId, setDetailOrderId] = useState<string | null>(null)
  const [cameraPermission, setCameraPermission] = useState<"idle" | "pending" | "granted" | "denied">("idle")
  const [scannerActive, setScannerActive] = useState(false)
  const [scannerKey, setScannerKey] = useState(0)
  const scannerRef = useRef<Html5Qrcode | null>(null)
  const scanLockedRef = useRef(false)
  const recentScanIdsRef = useRef<Map<string, number>>(new Map())
  const resumeTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null)

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
  }, [status?.status])

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
  }, [loadDevices, status?.status])

  useEffect(() => {
    return () => {
      if (resumeTimerRef.current) clearTimeout(resumeTimerRef.current)
    }
  }, [])

  const requestCameraAccess = useCallback(async () => {
    setError(null)
    setScannerActive(false)
    setScannerKey((k) => k + 1)
    setCameraPermission("pending")
    primeScanAudio()
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
    primeScanAudio()
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

  const scheduleResume = useCallback(
    (delayMs = SCAN_RESUME_DELAY_MS) => {
      if (resumeTimerRef.current) clearTimeout(resumeTimerRef.current)
      resumeTimerRef.current = setTimeout(() => {
        resumeScanning()
      }, delayMs)
    },
    [resumeScanning]
  )

  const pushScan = useCallback((scan: ScanResult) => {
    setScans((prev) => [scan, ...prev].slice(0, HISTORY_LIMIT))
  }, [])

  const updateScan = useCallback((key: string, patch: Partial<ScanResult>) => {
    setScans((prev) => prev.map((s) => (s.key === key ? { ...s, ...patch } : s)))
  }, [])

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

  const handleScanned = useCallback(
    async (code: string) => {
      const parsed = parseScan(code)
      if (!parsed) {
        playScanSound("error")
        setError("Ungültiger QR-Code")
        scheduleResume()
        return
      }

      const now = Date.now()
      const last = recentScanIdsRef.current.get(parsed.id)
      if (last && now - last < DUPLICATE_SCAN_WINDOW_MS) {
        scheduleResume()
        return
      }
      recentScanIdsRef.current.set(parsed.id, now)
      if (recentScanIdsRef.current.size > 32) {
        const cutoff = now - DUPLICATE_SCAN_WINDOW_MS * 4
        recentScanIdsRef.current.forEach((t, k) => {
          if (t < cutoff) recentScanIdsRef.current.delete(k)
        })
      }

      setError(null)
      const key = `${parsed.id}-${now}`
      pushScan({
        key,
        scanId: parsed.id,
        kind: parsed.kind,
        status: "loading",
        scannedAt: new Date(),
        lines: [],
      })

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
            const detail =
              j.detail ||
              j.message ||
              (res.status === 409
                ? "Maximale Einlösungen erreicht."
                : res.status === 410
                  ? "Kampagne nicht mehr aktiv."
                  : res.status === 404
                    ? "Kampagne nicht gefunden."
                    : `Fehler ${res.status}`)
            throw new Error(detail)
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
          const unredeemed = lines.filter((l) => !l.redemption)
          const allRedeemed = lines.length > 0 && unredeemed.length === 0
          const nextStatus: ScanStatus = lines.length === 0 ? "no-items" : allRedeemed ? "already-redeemed" : "success"
          updateScan(key, { status: nextStatus, orderId: data.orderId, lines })
          playScanSound(nextStatus === "success" ? "success" : "warning")
          scheduleResume()
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
        const referencedParentIds = new Set<string>()
        for (const l of stationLines) {
          if (l.parentLineId) referencedParentIds.add(l.parentLineId)
        }
        const bundleParents: Record<string, OrderLine> = {}
        for (const l of allLines) {
          if (referencedParentIds.has(l.id)) bundleParents[l.id] = l
        }
        const unredeemed = stationLines.filter((l) => !l.redemption)
        let nextStatus: ScanStatus
        if (stationLines.length === 0) {
          nextStatus = "no-items"
        } else if (unredeemed.length === 0) {
          nextStatus = "already-redeemed"
        } else {
          nextStatus = "success"
        }

        if (nextStatus === "success") {
          fireRedeem(orderData.id)
          const redeemedAt = new Date().toISOString()
          const optimisticLines = stationLines.map((l) =>
            l.redemption ? l : { ...l, redemption: { id: l.id, redeemedAt } }
          )
          updateScan(key, { status: nextStatus, orderId: orderData.id, lines: optimisticLines, bundleParents })
        } else {
          updateScan(key, { status: nextStatus, orderId: orderData.id, lines: stationLines, bundleParents })
        }

        playScanSound(nextStatus === "success" ? "success" : "warning")
        scheduleResume()
      } catch (e: unknown) {
        const message = e instanceof Error ? e.message : "Scan fehlgeschlagen"
        updateScan(key, { status: "error", message })
        playScanSound("error")
        scheduleResume()
      }
    },
    [bearerToken, fireRedeem, pushScan, scheduleResume, status, updateScan]
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

  const latest = scans[0] ?? null
  const history = scans.slice(1)

  return (
    <div className="bg-background text-foreground flex flex-1 flex-col gap-6 px-3 pt-6 pb-10 sm:px-6 sm:pt-8">
      <header className="mx-auto flex w-full max-w-xl items-center justify-between gap-3">
        <p className="min-w-0 truncate text-base font-semibold sm:text-lg">{status?.name ?? ""}</p>
        <div className="flex shrink-0 items-center gap-1.5">
          <HardRefreshButton />
          <FullscreenButton />
        </div>
      </header>

      {(!status || status.status !== "approved") && (
        <div className="flex flex-1 flex-col items-center">
          <p className="text-muted-foreground mt-2 max-w-md text-center">Scanne Abhol-QR-Codes hier</p>
          <div className="flex w-full flex-1 items-center justify-center">
            <PairingCodeDisplay deviceType="STATION" />
          </div>
        </div>
      )}

      {status?.status === "approved" && (
        <div className="mx-auto flex w-full flex-col items-center gap-8 sm:max-w-xl">
          {!scannerActive && (
            <div className="bg-card w-full rounded-2xl p-6">
              <div className="bg-primary/10 text-primary mx-auto mb-3 flex h-12 w-12 items-center justify-center rounded-full">
                <ScanLine className="h-6 w-6" aria-hidden />
              </div>
              <h2 className="font-primary text-center text-xl tracking-wide uppercase">Scanner starten</h2>
              <p className="text-muted-foreground mt-1 text-center text-sm">
                Erlaube den Kamerazugriff, um QR-Codes einzulösen.
              </p>

              <div className="mt-5 grid gap-3">
                {cameraPermission !== "granted" && (
                  <Button onClick={requestCameraAccess} disabled={cameraPermission === "pending"} className="w-full">
                    {cameraPermission === "pending" ? "Kamerazugriff wird angefragt…" : "Kamerazugriff erlauben"}
                  </Button>
                )}

                {cameraPermission === "granted" && (
                  <div className="grid w-full gap-2">
                    <Label htmlFor="camera" className="text-xs font-medium">
                      Kamera
                    </Label>
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
                          onClick={startScanner}
                          disabled={cameraPermission !== "granted" || !deviceId}
                          className="mt-1 w-full"
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
                  <div className="text-muted-foreground bg-card rounded-xl p-3 text-sm">
                    Zugriff abgelehnt. Bitte erlaube den Kamerazugriff in den Browser-Einstellungen und versuche es
                    erneut.
                  </div>
                )}
              </div>
            </div>
          )}

          {scannerActive && cameraPermission === "granted" && (
            <>
              <div className="flex w-full flex-col items-center">
                <div className="w-full max-w-[16rem] overflow-hidden rounded-2xl">
                  <Html5QrcodeScannerPlugin
                    key={`${deviceId}-${scannerKey}`}
                    cameraId={deviceId || undefined}
                    fps={10}
                    qrbox={200}
                    disableFlip={false}
                    maxWidthRem={16}
                    qrCodeSuccessCallback={handleDecoded}
                    onReady={handleScannerReady}
                    onStartError={handleScannerStartError}
                    onStop={handleScannerStop}
                  />
                </div>
                <p className="text-muted-foreground mt-3 inline-flex items-center gap-1.5 text-center text-xs">
                  <ScanLine className="h-3.5 w-3.5" aria-hidden />
                  QR-Code in den Rahmen halten
                </p>
              </div>

              <div className="flex w-full flex-col gap-8">
                {latest ? (
                  <ScanCard scan={latest} onOpenDetail={setDetailOrderId} />
                ) : (
                  <div className="text-muted-foreground bg-card rounded-2xl p-8 text-center text-sm">
                    Bereit zum Scannen — der nächste QR-Code wird automatisch eingelöst.
                  </div>
                )}

                {history.length > 0 && (
                  <div>
                    <div className="text-muted-foreground font-primary mb-3 px-1 text-xs tracking-[0.08em] uppercase">
                      Frühere Scans
                    </div>
                    <ul className="flex flex-col gap-1.5">
                      {history.map((s) => (
                        <ScanRow key={s.key} scan={s} onOpenDetail={setDetailOrderId} />
                      ))}
                    </ul>
                  </div>
                )}
              </div>
            </>
          )}

          {error && (
            <div role="alert" aria-live="assertive" className="text-destructive text-sm">
              {error}
            </div>
          )}
        </div>
      )}

      <OrderDetailDialog orderId={detailOrderId} bearerToken={bearerToken} onClose={() => setDetailOrderId(null)} />
    </div>
  )
}

function ScanCard({ scan, onOpenDetail }: { scan: ScanResult; onOpenDetail?: (orderId: string) => void }) {
  const meta = STATUS_STYLES[scan.status]
  const Icon = meta.icon
  const matched = scan.lines
  const isLoading = scan.status === "loading"
  const canOpenDetail = !!scan.orderId && !!onOpenDetail

  return (
    <div
      key={scan.key}
      className={`${meta.bg} animate-in fade-in slide-in-from-bottom-1 rounded-2xl duration-300 ease-out`}
    >
      <div className="flex items-center justify-between gap-3 px-5 pt-4 pb-3.5">
        <div className={`flex min-w-0 items-center gap-2 ${meta.text}`}>
          <Icon className={`h-[18px] w-[18px] shrink-0 ${isLoading ? "animate-spin" : ""}`} aria-hidden />
          <span className="font-primary mt-[3px] text-sm leading-none tracking-wide whitespace-nowrap uppercase">
            {meta.label}
          </span>
        </div>
        {canOpenDetail && (
          <button
            type="button"
            onClick={() => onOpenDetail?.(scan.orderId as string)}
            aria-label="Bestelldetails anzeigen"
            className="text-muted-foreground hover:bg-foreground/5 hover:text-foreground focus-visible:ring-ring/50 inline-flex h-7 w-7 shrink-0 items-center justify-center rounded-full transition-colors focus-visible:ring-2 focus-visible:outline-none"
          >
            <Info className="h-4 w-4" aria-hidden />
          </button>
        )}
      </div>

      {isLoading && (
        <div className="flex flex-col gap-3 px-5 pb-5">
          {[0, 1].map((i) => (
            <div key={i} className="rounded-xl border p-3">
              <div className="flex items-center gap-3">
                <div className="bg-foreground/5 h-16 w-16 shrink-0 animate-pulse rounded-[11px]" />
                <div className="flex-1 space-y-2">
                  <div className="bg-foreground/5 h-3.5 w-3/4 animate-pulse rounded" />
                  <div className="bg-foreground/5 h-3 w-1/2 animate-pulse rounded" />
                </div>
              </div>
            </div>
          ))}
        </div>
      )}

      {scan.status === "error" && (
        <p className="text-destructive px-5 pt-1 pb-5 text-sm">{scan.message || "Scan fehlgeschlagen"}</p>
      )}

      {scan.status === "no-items" && (
        <p className="text-muted-foreground px-5 pt-1 pb-5 text-sm">
          Diese Station bietet keine Artikel aus dieser Bestellung an.
        </p>
      )}

      {(scan.status === "success" || scan.status === "already-redeemed") && (
        <div className="flex flex-col gap-3 px-5 pb-5">
          {(() => {
            const items = consolidateLines(matched)
            const unredeemedItems = items.filter((it) => !it.redeemed)
            const redeemedItems = items.filter((it) => it.redeemed)
            return (
              <>
                {unredeemedItems.map((it) => (
                  <ConsolidatedItemCard key={it.productId} item={it} redeemed={false} />
                ))}
                {redeemedItems.length > 0 && scan.status === "success" && unredeemedItems.length > 0 && (
                  <div className="text-muted-foreground font-primary mt-1 text-[11px] tracking-[0.08em] uppercase">
                    Bereits eingelöst
                  </div>
                )}
                {redeemedItems.map((it) => (
                  <ConsolidatedItemCard key={it.productId} item={it} redeemed />
                ))}
              </>
            )
          })()}
        </div>
      )}
    </div>
  )
}

function ScanRow({ scan, onOpenDetail }: { scan: ScanResult; onOpenDetail?: (orderId: string) => void }) {
  const meta = STATUS_STYLES[scan.status]
  const itemCount = scan.lines.reduce((acc, l) => acc + l.quantity, 0)
  const canOpenDetail = !!scan.orderId && !!onOpenDetail
  return (
    <li className={`${meta.bg} flex items-center justify-between gap-3 rounded-xl px-4 py-3 text-sm`}>
      <div className="flex min-w-0 items-center gap-2.5">
        <span className={`h-2 w-2 shrink-0 rounded-full ${meta.dot}`} aria-hidden />
        <span
          className={`font-primary mt-[2px] shrink-0 text-[11px] leading-none tracking-wide uppercase ${meta.text}`}
        >
          {meta.label}
        </span>
        {itemCount > 0 && (
          <span className="text-muted-foreground truncate text-xs tabular-nums">{itemCount} Artikel</span>
        )}
      </div>
      <div className="text-muted-foreground inline-flex shrink-0 items-center gap-1 text-xs tabular-nums">
        <span>{scan.scannedAt.toLocaleTimeString("de-CH", { hour: "2-digit", minute: "2-digit" })}</span>
        {canOpenDetail && (
          <button
            type="button"
            onClick={() => onOpenDetail?.(scan.orderId as string)}
            aria-label="Bestelldetails anzeigen"
            className="text-muted-foreground hover:bg-foreground/5 hover:text-foreground focus-visible:ring-ring/50 ml-1 inline-flex h-7 w-7 items-center justify-center rounded-full transition-colors focus-visible:ring-2 focus-visible:outline-none"
          >
            <Info className="h-4 w-4" aria-hidden />
          </button>
        )}
      </div>
    </li>
  )
}

function ConsolidatedItemCard({ item, redeemed }: { item: ConsolidatedItem; redeemed: boolean }) {
  return (
    <div className="rounded-xl border p-3">
      <div className="flex items-center gap-3">
        {item.productImage ? (
          <div className="relative h-16 w-16 shrink-0 overflow-hidden rounded-[11px] bg-[#cec9c6]">
            <Image
              src={item.productImage}
              alt={item.title}
              fill
              sizes="64px"
              className="h-full w-full rounded-[11px] object-cover"
            />
          </div>
        ) : (
          <div className="bg-foreground/5 text-muted-foreground flex h-16 w-16 shrink-0 items-center justify-center rounded-[11px]">
            <PackageOpen className="h-5 w-5" aria-hidden />
          </div>
        )}
        <div className="min-w-0 flex-1">
          <p className="truncate font-medium">{item.title}</p>
          {redeemed && item.redemptionTime && (
            <div className="text-muted-foreground mt-0.5 text-xs tabular-nums">
              Eingelöst {new Date(item.redemptionTime).toLocaleTimeString("de-CH")}
            </div>
          )}
        </div>
        <div className="text-foreground/85 shrink-0 text-lg font-semibold tabular-nums">×{item.quantity}</div>
      </div>
    </div>
  )
}

function LineGroupCard({ group, redeemed }: { group: LineGroup; redeemed: boolean }) {
  const { parent, children } = group
  const isBundle = children.length > 0
  const redemptionTime = redeemed ? groupRedemptionTime(group) : undefined

  return (
    <div className="rounded-xl border p-3">
      <div className="flex items-center gap-3">
        {parent.productImage ? (
          <div className="relative h-16 w-16 shrink-0 overflow-hidden rounded-[11px] bg-[#cec9c6]">
            <Image
              src={parent.productImage}
              alt={parent.title}
              fill
              sizes="64px"
              className="h-full w-full rounded-[11px] object-cover"
            />
          </div>
        ) : (
          <div className="bg-foreground/5 text-muted-foreground flex h-16 w-16 shrink-0 items-center justify-center rounded-[11px]">
            <PackageOpen className="h-5 w-5" aria-hidden />
          </div>
        )}
        <div className="min-w-0 flex-1">
          <p className="truncate font-medium">{parent.title}</p>
          {isBundle ? (
            <div className="mt-1 flex flex-row flex-wrap gap-1.5">
              {children.map((c) => (
                <span key={c.id} className="text-muted-foreground border-border rounded-lg border px-2 py-0.5 text-xs">
                  {c.menuSlotName ?? "Option"}: {c.title}
                </span>
              ))}
            </div>
          ) : (
            parent.menuSlotName && (
              <div className="text-muted-foreground mt-0.5 truncate text-xs">{parent.menuSlotName}</div>
            )
          )}
          {redeemed && redemptionTime && (
            <div className="text-muted-foreground mt-0.5 text-xs tabular-nums">
              Eingelöst {new Date(redemptionTime).toLocaleTimeString("de-CH")}
            </div>
          )}
        </div>
        <div className="text-foreground/85 shrink-0 text-lg font-semibold tabular-nums">×{parent.quantity}</div>
      </div>
    </div>
  )
}

type DetailOrder = {
  id: string
  lines?: OrderLine[]
}

function OrderDetailDialog({
  orderId,
  bearerToken,
  onClose,
}: {
  orderId: string | null
  bearerToken: string
  onClose: () => void
}) {
  const [order, setOrder] = useState<DetailOrder | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (!orderId) {
      setOrder(null)
      setError(null)
      return
    }
    let cancelled = false
    setLoading(true)
    setError(null)
    setOrder(null)
    ;(async () => {
      try {
        const res = await fetch(`/api/v1/orders/${encodeURIComponent(orderId)}`, {
          headers: { Authorization: `Bearer ${bearerToken}` },
        })
        if (!res.ok) {
          const j = (await res.json().catch(() => ({}))) as { detail?: string }
          throw new Error(j.detail || `Fehler ${res.status}`)
        }
        const data = (await res.json()) as DetailOrder
        if (!cancelled) setOrder(data)
      } catch (e) {
        if (!cancelled) setError(e instanceof Error ? e.message : "Bestellung konnte nicht geladen werden.")
      } finally {
        if (!cancelled) setLoading(false)
      }
    })()
    return () => {
      cancelled = true
    }
  }, [orderId, bearerToken])

  const allLines = order?.lines ?? []
  const dialogBundleParents: Record<string, OrderLine> = {}
  for (const l of allLines) {
    if (l.lineType === "bundle") dialogBundleParents[l.id] = l
  }
  const dialogGroups = groupLines(
    allLines.filter((l) => l.lineType !== "bundle"),
    dialogBundleParents
  )

  return (
    <Dialog open={!!orderId} onOpenChange={(open) => !open && onClose()}>
      <DialogContent className="max-h-[85vh] overflow-y-auto sm:max-w-md">
        <DialogHeader>
          <DialogTitle className="font-primary text-center text-xl tracking-wide uppercase">Bestelldetails</DialogTitle>
        </DialogHeader>

        {loading && (
          <div className="flex items-center justify-center py-12">
            <Loader2 className="text-muted-foreground h-5 w-5 animate-spin" aria-hidden />
          </div>
        )}

        {error && !loading && <p className="text-destructive py-4 text-sm">{error}</p>}

        {order && !loading && (
          <>
            {dialogGroups.length === 0 ? (
              <p className="text-muted-foreground py-4 text-sm">Keine Artikel in dieser Bestellung.</p>
            ) : (
              <ul className="flex flex-col gap-3">
                {dialogGroups.map((g) => (
                  <li key={g.key}>
                    <LineGroupCard group={g} redeemed={false} />
                  </li>
                ))}
              </ul>
            )}
          </>
        )}
      </DialogContent>
    </Dialog>
  )
}
