"use client"

import { ArrowLeft, Check, Copy, Loader2, RefreshCw } from "lucide-react"
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

type SlotState = "redeemed" | "reserved" | "available"

const SLOT_STATE_LABEL: Record<SlotState, string> = {
  redeemed: "eingelöst",
  reserved: "reserviert",
  available: "verfügbar",
}

const SLOT_STATE_CLASS: Record<SlotState, string> = {
  redeemed: "bg-emerald-50 text-emerald-900",
  reserved: "bg-amber-50 text-amber-900",
  available: "bg-muted/60 text-muted-foreground",
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

  if (!firstLoadDone) {
    return (
      <div className="flex flex-col gap-6 px-6 md:px-8 lg:px-10">
        <Card>
          <CardHeader>
            <Skeleton className="h-4 w-32" />
            <Skeleton className="h-7 w-56" />
          </CardHeader>
          <CardContent>
            <div className="grid gap-4 sm:grid-cols-4">
              {[0, 1, 2, 3].map((i) => (
                <Skeleton key={i} className="h-12 w-full" />
              ))}
            </div>
          </CardContent>
        </Card>
        <Skeleton className="h-40 w-full rounded-xl" />
      </div>
    )
  }

  if (error && !detail) {
    return (
      <div className="flex flex-col gap-3 px-6 py-10 md:px-8 lg:px-10">
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
  const progressPct = detail.totalSlots > 0 ? Math.round((detail.redeemedSlots / detail.totalSlots) * 100) : 0

  return (
    <div className="flex flex-col gap-6 px-6 md:px-8 lg:px-10">
      {/* Header */}
      <Card>
        <CardHeader className="flex-col items-start gap-3 sm:flex-row sm:items-start sm:justify-between">
          <div className="flex min-w-0 flex-col gap-2">
            <Link
              href="/admin/staff-meals"
              className="text-muted-foreground hover:text-foreground -ml-1 flex w-fit items-center gap-1 text-xs"
            >
              <ArrowLeft className="size-3" aria-hidden />
              Mitarbeiter-Essen
            </Link>
            <div className="flex flex-wrap items-center gap-2">
              <CardTitle className="break-words">{detail.name}</CardTitle>
              <span className={`rounded-full px-2 py-0.5 text-xs font-medium ${STATUS_CLASS[detail.status]}`}>
                {STATUS_LABEL[detail.status]}
              </span>
            </div>
          </div>
          <div className="flex shrink-0 flex-wrap gap-2">
            {detail.status !== "ended" && (
              <Button variant="outline" onClick={() => setEndConfirmOpen(true)}>
                Beenden
              </Button>
            )}
          </div>
        </CardHeader>
        <CardContent className="flex flex-col gap-5 text-sm">
          <div className="grid grid-cols-2 gap-4 sm:grid-cols-4">
            <StatCell label="Eingelöst" value={`${detail.redeemedSlots} / ${detail.totalSlots}`} />
            <StatCell label="Reserviert" value={String(detail.reservedSlots)} />
            <StatCell label="Gültig bis" value={detail.validUntil ? formatShort(detail.validUntil) : "—"} />
            <StatCell label="Erstellt" value={formatShort(detail.createdAt)} />
          </div>
          <div className="flex flex-col gap-1.5">
            <div className="text-muted-foreground flex items-baseline justify-between text-xs tabular-nums">
              <span>Einlösung</span>
              <span>{progressPct}%</span>
            </div>
            <div className="bg-muted relative h-1.5 overflow-hidden rounded-full">
              <div
                className="bg-primary absolute inset-y-0 left-0 rounded-full transition-[width] duration-300"
                style={{ width: `${progressPct}%` }}
              />
            </div>
          </div>
          {error && <div className="bg-destructive/10 text-destructive rounded-xl p-3 text-sm">{error}</div>}
        </CardContent>
      </Card>

      {/* Share */}
      <Card className="bg-primary/5 border-primary/20">
        <CardHeader>
          <CardTitle className="text-base">An Mitarbeiter weitergeben</CardTitle>
        </CardHeader>
        <CardContent className="flex flex-col gap-5">
          <div className="grid gap-4 md:grid-cols-[auto_1fr] md:items-start">
            <div className="flex flex-col gap-2">
              <div className="text-muted-foreground text-xs font-medium tracking-wide uppercase">Zugangscode</div>
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
              <div className="text-muted-foreground text-xs font-medium tracking-wide uppercase">Claim-URL</div>
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
              <p className="text-muted-foreground text-xs leading-relaxed">
                Sende Link + Code an die Mitarbeiter. Jeder wählt auf der Seite einen QR-Code und zeigt ihn an der
                Station.
              </p>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Products */}
      <Card>
        <CardHeader>
          <CardTitle className="text-base">Produkte pro Mitarbeiter</CardTitle>
        </CardHeader>
        <CardContent>
          {detail.products.length === 0 ? (
            <p className="text-muted-foreground text-sm">Keine Produkte.</p>
          ) : (
            <ul className="flex flex-col gap-1.5">
              {detail.products.map((p) => (
                <li key={p.productId} className="bg-muted/40 flex items-center justify-between rounded-lg px-3 py-2.5">
                  <span className="truncate">{p.productName}</span>
                  <span className="text-muted-foreground ml-3 shrink-0 tabular-nums">×{p.quantity}</span>
                </li>
              ))}
            </ul>
          )}
        </CardContent>
      </Card>

      {/* Slots */}
      <Card>
        <CardHeader>
          <CardTitle className="text-base">Slots ({detail.slots.length})</CardTitle>
        </CardHeader>
        <CardContent className="p-0 sm:p-6">
          <div className="overflow-x-auto">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead className="w-12 pl-6 sm:pl-4">#</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead>Eingelöst</TableHead>
                  <TableHead>Reserviert bis</TableHead>
                  <TableHead className="pr-6 sm:pr-4">Order</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {detail.slots.map((s, idx) => {
                  const nowMs = Date.now()
                  const reserved = s.reservedUntil && new Date(s.reservedUntil).getTime() > nowMs
                  const state: SlotState = s.isRedeemed ? "redeemed" : reserved ? "reserved" : "available"
                  return (
                    <TableRow key={s.id}>
                      <TableCell className="text-muted-foreground pl-6 tabular-nums sm:pl-4">{idx + 1}</TableCell>
                      <TableCell>
                        <span className={`rounded-full px-2 py-0.5 text-xs font-medium ${SLOT_STATE_CLASS[state]}`}>
                          {SLOT_STATE_LABEL[state]}
                        </span>
                      </TableCell>
                      <TableCell className="text-muted-foreground text-sm tabular-nums">
                        {s.isRedeemed && s.redeemedAt ? formatShort(s.redeemedAt) : "—"}
                      </TableCell>
                      <TableCell className="text-muted-foreground text-sm tabular-nums">
                        {reserved && s.reservedUntil
                          ? new Date(s.reservedUntil).toLocaleTimeString("de-CH", {
                              hour: "2-digit",
                              minute: "2-digit",
                            })
                          : "—"}
                      </TableCell>
                      <TableCell className="text-muted-foreground pr-6 font-mono text-xs sm:pr-4">
                        {s.orderId.slice(0, 8)}
                      </TableCell>
                    </TableRow>
                  )
                })}
              </TableBody>
            </Table>
          </div>
        </CardContent>
      </Card>

      <AlertDialog open={endConfirmOpen} onOpenChange={setEndConfirmOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Mitarbeiter-Essen beenden?</AlertDialogTitle>
            <AlertDialogDescription>
              Alle noch nicht eingelösten QR-Codes werden storniert. Diese Aktion kann nicht rückgängig gemacht werden.
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
              Der aktuelle Link wird ungültig. Alle Mitarbeiter benötigen den neuen Link, um QR-Codes zu erhalten.
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
      <div className="text-muted-foreground text-xs font-medium tracking-wide uppercase">{label}</div>
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
