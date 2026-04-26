"use client"

import { AlertTriangle, ArrowLeft, Check, Copy, Loader2, Printer, RefreshCw } from "lucide-react"
import Link from "next/link"
import { useParams } from "next/navigation"
import { useCallback, useEffect, useState } from "react"
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Skeleton } from "@/components/ui/skeleton"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { useAuthorizedFetch } from "@/hooks/use-authorized-fetch"
import { getCSRFToken } from "@/lib/csrf"
import { readErrorMessage } from "@/lib/http"
import type { VolunteerCampaignDetail, VolunteerCampaignStatus } from "@/types/volunteer"

const STATUS_LABEL: Record<VolunteerCampaignStatus, string> = {
  draft: "Entwurf",
  active: "Aktiv",
  ended: "Beendet",
}

const STATUS_CLASS: Record<VolunteerCampaignStatus, string> = {
  draft: "bg-amber-100 text-amber-900",
  active: "bg-emerald-100 text-emerald-900",
  ended: "bg-muted text-muted-foreground",
}

export default function AdminStaffMealDetailPage() {
  const params = useParams<{ id: string }>()
  const id = params?.id
  const fetchAuth = useAuthorizedFetch()
  const [detail, setDetail] = useState<VolunteerCampaignDetail | null>(null)
  const [firstLoadDone, setFirstLoadDone] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [copiedUrl, setCopiedUrl] = useState(false)
  const [copiedCode, setCopiedCode] = useState(false)
  const [rotating, setRotating] = useState(false)
  const [endConfirmOpen, setEndConfirmOpen] = useState(false)
  const [rotateConfirmOpen, setRotateConfirmOpen] = useState(false)

  const [editingMax, setEditingMax] = useState(false)
  const [maxDraft, setMaxDraft] = useState<number | "">(0)
  const [savingMax, setSavingMax] = useState(false)
  const [maxError, setMaxError] = useState<string | null>(null)

  const [printOpen, setPrintOpen] = useState(false)
  const [printCount, setPrintCount] = useState<number | "">(30)

  const load = useCallback(async () => {
    if (!id) return
    try {
      const res = await fetchAuth(`/api/v1/staff-meals/${encodeURIComponent(id)}`)
      if (!res.ok) throw new Error(await readErrorMessage(res))
      const j = (await res.json()) as VolunteerCampaignDetail
      setDetail(j)
      setError(null)
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : "Laden fehlgeschlagen")
    } finally {
      setFirstLoadDone(true)
    }
  }, [fetchAuth, id])

  useEffect(() => {
    load()
    const t = window.setInterval(load, 5000)
    return () => window.clearInterval(t)
  }, [load])

  async function endCampaign() {
    if (!id) return
    try {
      const res = await fetchAuth(`/api/v1/staff-meals/${encodeURIComponent(id)}/end`, {
        method: "POST",
        headers: { "X-CSRF": getCSRFToken() || "" },
      })
      if (!res.ok && res.status !== 204) throw new Error(await readErrorMessage(res))
      await load()
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : "Beenden fehlgeschlagen")
    }
  }

  async function rotateToken() {
    if (!id) return
    setRotating(true)
    try {
      const res = await fetchAuth(`/api/v1/staff-meals/${encodeURIComponent(id)}/rotate-token`, {
        method: "POST",
        headers: { "X-CSRF": getCSRFToken() || "" },
      })
      if (!res.ok) throw new Error(await readErrorMessage(res))
      await load()
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : "Link-Rotation fehlgeschlagen")
    } finally {
      setRotating(false)
    }
  }

  async function saveMax() {
    if (!id || !detail) return
    setMaxError(null)
    const maxValue = maxDraft === "" ? 0 : maxDraft
    if (maxValue < detail.redemptionCount) {
      setMaxError(`Nicht unter aktueller Einlösungszahl (${detail.redemptionCount}).`)
      return
    }
    if (maxValue < 1) {
      setMaxError("Muss mindestens 1 sein.")
      return
    }
    setSavingMax(true)
    try {
      const res = await fetchAuth(`/api/v1/staff-meals/${encodeURIComponent(id)}`, {
        method: "PATCH",
        headers: { "Content-Type": "application/json", "X-CSRF": getCSRFToken() || "" },
        body: JSON.stringify({
          name: detail.name,
          validFrom: detail.validFrom ?? undefined,
          validUntil: detail.validUntil ?? undefined,
          status: detail.status,
          maxRedemptions: maxValue,
        }),
      })
      if (!res.ok) throw new Error(await readErrorMessage(res))
      setEditingMax(false)
      await load()
    } catch (e: unknown) {
      setMaxError(e instanceof Error ? e.message : "Speichern fehlgeschlagen")
    } finally {
      setSavingMax(false)
    }
  }

  function copyClaimUrl() {
    if (!detail) return
    const url = `${window.location.origin}/claim/${detail.claimToken}`
    void navigator.clipboard.writeText(url)
    setCopiedUrl(true)
    window.setTimeout(() => setCopiedUrl(false), 1800)
  }

  function copyAccessCode() {
    if (!detail) return
    void navigator.clipboard.writeText(detail.accessCode)
    setCopiedCode(true)
    window.setTimeout(() => setCopiedCode(false), 1800)
  }

  function openPrint() {
    setPrintCount(30)
    setPrintOpen(true)
  }

  function submitPrint() {
    if (!id) return
    const raw = printCount === "" ? 0 : printCount
    const count = Math.max(1, Math.min(500, raw || 0))
    const url = `/api/v1/staff-meals/${encodeURIComponent(id)}/print.pdf?count=${count}`
    window.open(url, "_blank")
    setPrintOpen(false)
  }

  if (!firstLoadDone) {
    return (
      <div className="space-y-4 pt-4">
        <Skeleton className="h-8 w-56" />
        <Skeleton className="h-32 w-full rounded-2xl" />
        <Skeleton className="h-40 w-full rounded-2xl" />
      </div>
    )
  }

  if (error && !detail) {
    return (
      <div className="space-y-3 pt-4">
        <Link
          href="/admin/staff-meals"
          className="text-muted-foreground hover:text-foreground flex items-center gap-1.5 text-sm"
        >
          <ArrowLeft className="size-4" aria-hidden />
          Mitarbeiter-Essen
        </Link>
        <p className="text-destructive">{error}</p>
      </div>
    )
  }
  if (!detail) return null

  const claimUrl = typeof window !== "undefined" ? `${window.location.origin}/claim/${detail.claimToken}` : ""
  const progressPct =
    detail.maxRedemptions > 0 ? Math.min(100, Math.round((detail.redemptionCount / detail.maxRedemptions) * 100)) : 0
  const expired = detail.validUntil ? new Date(detail.validUntil).getTime() < Date.now() : false
  const printWarning = detail.status === "ended" || expired

  return (
    <div className="space-y-4 pt-4">
      {/* Header row */}
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div className="flex min-w-0 flex-col gap-1">
          <Link
            href="/admin/staff-meals"
            className="text-muted-foreground hover:text-foreground -ml-1 flex w-fit items-center gap-1 text-sm"
          >
            <ArrowLeft className="size-3.5" aria-hidden />
            Mitarbeiter-Essen
          </Link>
          <div className="flex flex-wrap items-center gap-2">
            <h1 className="font-primary truncate text-2xl">{detail.name}</h1>
            <span className={`rounded-full px-2 py-0.5 text-xs font-medium ${STATUS_CLASS[detail.status]}`}>
              {STATUS_LABEL[detail.status]}
            </span>
          </div>
        </div>
        <div className="flex shrink-0 flex-wrap gap-2">
          <Button onClick={openPrint}>
            <Printer className="size-4" aria-hidden />
            Drucken
          </Button>
          {detail.status !== "ended" && (
            <Button variant="outline" onClick={() => setEndConfirmOpen(true)}>
              Beenden
            </Button>
          )}
        </div>
      </div>

      {error && <div className="text-destructive bg-destructive/10 rounded-xl px-3 py-2 text-sm">{error}</div>}

      {/* Stats */}
      <Card className="rounded-2xl">
        <CardContent className="flex flex-col gap-5 p-6 text-sm">
          <div className="grid grid-cols-2 gap-4 sm:grid-cols-4">
            <StatCell label="Eingelöst" value={`${detail.redemptionCount} / ${detail.maxRedemptions}`} />
            <div className="flex flex-col gap-1">
              <div className="text-muted-foreground text-xs">Maximum</div>
              {editingMax ? (
                <div className="flex items-center gap-1">
                  <Input
                    type="number"
                    value={maxDraft}
                    min={detail.redemptionCount || 1}
                    onChange={(e) => {
                      const raw = e.target.value
                      setMaxDraft(raw === "" ? "" : parseInt(raw, 10))
                    }}
                    className="h-8 w-20"
                  />
                  <Button size="sm" onClick={saveMax} disabled={savingMax}>
                    {savingMax ? <Loader2 className="size-3.5 animate-spin" aria-hidden /> : "OK"}
                  </Button>
                  <Button
                    size="sm"
                    variant="ghost"
                    onClick={() => {
                      setEditingMax(false)
                      setMaxError(null)
                    }}
                    disabled={savingMax}
                  >
                    Abbr.
                  </Button>
                </div>
              ) : (
                <button
                  type="button"
                  onClick={() => {
                    setMaxDraft(detail.maxRedemptions)
                    setMaxError(null)
                    setEditingMax(true)
                  }}
                  className="hover:text-foreground text-left text-base font-semibold tabular-nums underline-offset-4 hover:underline"
                >
                  {detail.maxRedemptions}
                </button>
              )}
              {maxError && <span className="text-destructive text-xs">{maxError}</span>}
            </div>
            <StatCell label="Gültig bis" value={detail.validUntil ? formatShort(detail.validUntil) : "—"} />
            <StatCell label="Erstellt" value={formatShort(detail.createdAt)} />
          </div>
          <div className="flex flex-col gap-1.5">
            <div className="text-muted-foreground flex items-baseline justify-between text-xs tabular-nums">
              <span>Einlösung</span>
              <span>{progressPct}%</span>
            </div>
            <div className="bg-muted-foreground/20 relative h-1.5 overflow-hidden rounded-full">
              <div
                className="bg-primary absolute inset-y-0 left-0 rounded-full transition-[width] duration-300"
                style={{ width: `${progressPct}%` }}
              />
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Share */}
      <Card className="rounded-2xl">
        <CardHeader>
          <CardTitle>An Mitarbeiter weitergeben</CardTitle>
          <p className="text-muted-foreground text-sm">
            Teile Link und Code mit dem Team. Der geteilte QR-Code kann bis zu {detail.maxRedemptions}-mal eingelöst
            werden.
          </p>
        </CardHeader>
        <CardContent className="flex flex-col gap-5">
          <div className="grid gap-4 md:grid-cols-[auto_1fr] md:items-start">
            <div className="flex flex-col gap-2">
              <div className="text-muted-foreground text-xs">Zugangscode</div>
              <button
                type="button"
                onClick={copyAccessCode}
                className="group bg-background hover:bg-accent/40 focus-visible:ring-ring/40 flex items-center gap-3 rounded-xl border px-4 py-3 text-left transition focus-visible:ring-2 focus-visible:outline-none active:scale-[0.99]"
                aria-label="Code kopieren"
              >
                <span className="font-mono text-3xl font-semibold tracking-[0.3em]">{detail.accessCode}</span>
                <span className="text-muted-foreground group-hover:text-foreground transition">
                  {copiedCode ? (
                    <Check className="size-4 text-emerald-600" aria-hidden />
                  ) : (
                    <Copy className="size-4" aria-hidden />
                  )}
                </span>
              </button>
            </div>
            <div className="flex min-w-0 flex-col gap-2">
              <div className="text-muted-foreground text-xs">Claim-URL</div>
              <div className="bg-background flex items-center gap-1 rounded-xl border p-1 pl-3">
                <code className="min-w-0 flex-1 truncate text-sm">{claimUrl}</code>
                <Button variant="ghost" size="sm" onClick={copyClaimUrl} aria-label="URL kopieren" className="shrink-0">
                  {copiedUrl ? (
                    <Check className="size-4 text-emerald-600" aria-hidden />
                  ) : (
                    <Copy className="size-4" aria-hidden />
                  )}
                  <span className="ml-1 hidden sm:inline">{copiedUrl ? "Kopiert" : "Kopieren"}</span>
                </Button>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => setRotateConfirmOpen(true)}
                  disabled={rotating}
                  aria-label="Neuen Link generieren"
                  className="shrink-0"
                >
                  <RefreshCw className={`size-4 ${rotating ? "animate-spin" : ""}`} aria-hidden />
                </Button>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Products */}
      <Card className="rounded-2xl">
        <CardHeader>
          <CardTitle>Produkte pro Einlösung</CardTitle>
        </CardHeader>
        <CardContent>
          {detail.products.length === 0 ? (
            <p className="text-muted-foreground text-sm">Keine Produkte.</p>
          ) : (
            <ul className="flex flex-col gap-1.5">
              {detail.products.map((p) => (
                <li key={p.productId} className="bg-muted/40 flex items-center justify-between rounded-xl px-3 py-2.5">
                  <span className="truncate">{p.productName}</span>
                  <span className="text-muted-foreground ml-3 shrink-0 tabular-nums">×{p.quantity}</span>
                </li>
              ))}
            </ul>
          )}
        </CardContent>
      </Card>

      {/* Redemptions */}
      <Card className="rounded-2xl">
        <CardHeader>
          <CardTitle>Einlösungen ({detail.redemptionCount})</CardTitle>
        </CardHeader>
        <CardContent className="p-0 sm:px-6 sm:pb-6">
          <div className="overflow-x-auto">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead className="w-12 pl-6 sm:pl-4">#</TableHead>
                  <TableHead>Zeitpunkt</TableHead>
                  <TableHead className="pr-6 sm:pr-4">Order</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {detail.redemptions.length === 0 ? (
                  <TableRow>
                    <TableCell colSpan={3} className="text-muted-foreground py-8 text-center text-sm">
                      Noch keine Einlösungen.
                    </TableCell>
                  </TableRow>
                ) : (
                  detail.redemptions.map((r, idx) => (
                    <TableRow key={r.id}>
                      <TableCell className="text-muted-foreground pl-6 tabular-nums sm:pl-4">
                        {detail.redemptionCount - idx}
                      </TableCell>
                      <TableCell className="text-muted-foreground text-sm tabular-nums">
                        {formatShort(r.createdAt)}
                      </TableCell>
                      <TableCell className="text-muted-foreground pr-6 font-mono text-xs sm:pr-4">
                        {r.orderId.slice(0, 8)}
                      </TableCell>
                    </TableRow>
                  ))
                )}
              </TableBody>
            </Table>
          </div>
        </CardContent>
      </Card>

      {/* Print dialog */}
      <Dialog open={printOpen} onOpenChange={setPrintOpen}>
        <DialogContent className="max-w-md">
          <DialogHeader>
            <DialogTitle>QR-Slips drucken</DialogTitle>
          </DialogHeader>
          <div className="flex flex-col gap-4">
            {printWarning && (
              <div className="flex items-start gap-2 rounded-xl bg-amber-50 p-3 text-sm text-amber-900">
                <AlertTriangle className="mt-0.5 size-4 shrink-0" aria-hidden />
                <div>
                  <p className="font-medium">Kampagne nicht mehr einlösbar</p>
                  <p className="mt-0.5 text-xs">Die gedruckten Slips werden an der Station abgelehnt.</p>
                </div>
              </div>
            )}
            <div className="grid gap-2">
              <Label htmlFor="print-count">Anzahl Slips</Label>
              <Input
                id="print-count"
                type="number"
                min={1}
                max={500}
                value={printCount}
                onChange={(e) => {
                  const raw = e.target.value
                  setPrintCount(raw === "" ? "" : parseInt(raw, 10))
                }}
                autoFocus
              />
              <p className="text-muted-foreground text-xs">
                5 Slips pro Reihe auf A4. Alle enthalten den gleichen QR-Code.
              </p>
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setPrintOpen(false)}>
              Abbrechen
            </Button>
            <Button onClick={submitPrint} disabled={printCount === "" || printCount < 1 || printCount > 500}>
              PDF öffnen
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <AlertDialog open={endConfirmOpen} onOpenChange={setEndConfirmOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Mitarbeiter-Essen beenden?</AlertDialogTitle>
            <AlertDialogDescription>
              Der QR-Code kann nach dem Beenden nicht mehr eingelöst werden. Diese Aktion kann nicht rückgängig gemacht
              werden.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Abbrechen</AlertDialogCancel>
            <AlertDialogAction
              onClick={async () => {
                setEndConfirmOpen(false)
                await endCampaign()
              }}
            >
              Beenden
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      <AlertDialog open={rotateConfirmOpen} onOpenChange={setRotateConfirmOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Neuen Link generieren?</AlertDialogTitle>
            <AlertDialogDescription>
              Der aktuelle Link und bereits gedruckte QR-Codes werden ungültig. Alle Mitarbeiter benötigen den neuen
              Link.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Abbrechen</AlertDialogCancel>
            <AlertDialogAction
              onClick={async () => {
                setRotateConfirmOpen(false)
                await rotateToken()
              }}
            >
              {rotating ? (
                <>
                  <Loader2 className="size-4 animate-spin" aria-hidden />
                  Generiere …
                </>
              ) : (
                "Neu generieren"
              )}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}

function StatCell({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex flex-col gap-1">
      <div className="text-muted-foreground text-xs">{label}</div>
      <div className="text-base font-semibold tabular-nums">{value}</div>
    </div>
  )
}

function formatShort(iso: string): string {
  const d = new Date(iso)
  return (
    d.toLocaleDateString("de-CH", { day: "2-digit", month: "2-digit", year: "2-digit" }) +
    " " +
    d.toLocaleTimeString("de-CH", { hour: "2-digit", minute: "2-digit" })
  )
}
