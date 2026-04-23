"use client"

import { ArrowLeft, Check, CheckCircle2, Loader2, Timer } from "lucide-react"
import Image from "next/image"
import Link from "next/link"
import { useParams, useRouter } from "next/navigation"
import { useCallback, useEffect, useMemo, useRef, useState } from "react"
import QRCode from "@/components/qrcode"
import { Button } from "@/components/ui/button"
import { readErrorMessage } from "@/lib/http"
import type { ClaimSlotDetail } from "@/types/volunteer"

export default function ClaimSlotDetailPage() {
  const params = useParams<{ token: string; slotId: string }>()
  const router = useRouter()
  const token = params?.token
  const slotId = params?.slotId

  const [slot, setSlot] = useState<ClaimSlotDetail | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [firstLoadDone, setFirstLoadDone] = useState(false)
  const [releasing, setReleasing] = useState(false)
  const [now, setNow] = useState(() => Date.now())
  const footerRef = useRef<HTMLDivElement>(null)
  const [footerHeight, setFooterHeight] = useState(0)

  const load = useCallback(async () => {
    if (!token || !slotId) return
    try {
      const res = await fetch(`/api/v1/claim/${encodeURIComponent(token)}/slots/${encodeURIComponent(slotId)}`, {
        credentials: "include",
      })
      if (res.status === 401) {
        router.replace(`/claim/${token}`)
        return
      }
      if (res.status === 404 || res.status === 409) {
        setError("Dieser QR-Code ist nicht mehr verfügbar.")
        return
      }
      if (!res.ok) throw new Error(await readErrorMessage(res))
      const j = (await res.json()) as ClaimSlotDetail
      setSlot(j)
      setError(null)
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : "Laden fehlgeschlagen")
    } finally {
      setFirstLoadDone(true)
    }
  }, [token, slotId, router])

  useEffect(() => {
    load()
    const t = window.setInterval(load, 4000)
    return () => window.clearInterval(t)
  }, [load])

  useEffect(() => {
    const t = window.setInterval(() => setNow(Date.now()), 1000)
    return () => window.clearInterval(t)
  }, [])

  useEffect(() => {
    const el = footerRef.current
    if (!el) return
    const update = () => setFooterHeight(el.offsetHeight || 0)
    update()
    let ro: ResizeObserver | null = null
    if (typeof ResizeObserver !== "undefined") {
      ro = new ResizeObserver(update)
      ro.observe(el)
    }
    window.addEventListener("resize", update)
    return () => {
      window.removeEventListener("resize", update)
      ro?.disconnect()
    }
  }, [])

  async function releaseSlot() {
    if (!token || !slotId) return
    setReleasing(true)
    try {
      const res = await fetch(
        `/api/v1/claim/${encodeURIComponent(token)}/slots/${encodeURIComponent(slotId)}/release`,
        { method: "POST", credentials: "include" }
      )
      if (!res.ok && res.status !== 204) {
        setError(await readErrorMessage(res))
        return
      }
      router.replace(`/claim/${token}`)
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : "Freigeben fehlgeschlagen")
    } finally {
      setReleasing(false)
    }
  }

  const countdown = useMemo(() => {
    if (!slot?.reservedUntil) return null
    const ms = new Date(slot.reservedUntil).getTime() - now
    if (ms <= 0) return null
    const mins = Math.floor(ms / 60000)
    const secs = Math.floor((ms % 60000) / 1000)
    return `${mins}:${String(secs).padStart(2, "0")}`
  }, [slot?.reservedUntil, now])

  if (!token || !slotId) return null

  // Loading: keep consistent with order detail page — animate-pulse placeholder
  if (!firstLoadDone) {
    return (
      <div className="mx-auto max-w-xl p-4">
        <div className="mb-2 h-8 w-40 animate-pulse rounded bg-gray-100" />
        <div className="mb-6 h-4 w-64 animate-pulse rounded bg-gray-100" />
        <div className="mx-auto h-[260px] w-[260px] animate-pulse rounded-[11px] border-2 bg-gray-100" />
      </div>
    )
  }

  if (error && !slot) {
    return (
      <div className="mx-auto max-w-xl p-4" style={{ paddingBottom: footerHeight ? footerHeight + 16 : 16 }}>
        <h1 className="mb-2 text-2xl font-semibold">QR-Code nicht verfügbar</h1>
        <p className="text-muted-foreground mb-6 text-sm">Diese Mahlzeit ist bereits eingelöst oder abgelaufen.</p>
        <div className="flex flex-col items-center gap-3 py-8 text-center">
          <Timer className="text-muted-foreground size-9" aria-hidden />
          <p className="text-destructive text-sm">{error}</p>
        </div>

        <div ref={footerRef} className="fixed inset-x-0 bottom-0 z-50 mx-auto max-w-xl p-4">
          <Button
            asChild
            variant="outline"
            className="rounded-pill h-12 w-full bg-white text-base font-medium text-black"
          >
            <Link href={`/claim/${token}`}>
              <ArrowLeft style={{ width: 20, height: 20 }} aria-hidden />
              Zurück zur Liste
            </Link>
          </Button>
        </div>
      </div>
    )
  }

  if (!slot) return null

  const isRedeemed = slot.isRedeemed
  const totalLines = slot.lines.length

  return (
    <div
      className="mx-auto flex max-w-xl flex-col p-4"
      style={{ paddingBottom: footerHeight ? footerHeight + 16 : 16 }}
    >
      <h1 className="mb-2 text-2xl font-semibold">{isRedeemed ? "Eingelöst" : "Dein QR-Code"}</h1>
      <p className="text-muted-foreground mb-6 text-sm">
        {isRedeemed ? "Guten Appetit!" : "Zeig diesen QR-Code an der Station vor."}
      </p>

      {isRedeemed ? (
        <div
          className="mx-auto flex w-full max-w-[260px] flex-col items-center gap-3 rounded-[11px] border-2 border-emerald-200 bg-emerald-50 p-8 text-center"
          style={{ animation: "claimPop 320ms cubic-bezier(0.2, 0.8, 0.2, 1) both" }}
        >
          <div
            className="flex h-14 w-14 items-center justify-center rounded-full bg-emerald-500 text-white"
            style={{ animation: "checkPop 380ms cubic-bezier(0.2, 0.9, 0.3, 1.2) 120ms both" }}
          >
            <CheckCircle2 className="size-7" aria-hidden />
          </div>
          <div className="text-emerald-900">
            <p className="font-semibold">Guten Appetit!</p>
            {slot.redeemedAt && (
              <p className="mt-0.5 text-xs text-emerald-700 tabular-nums">
                {new Date(slot.redeemedAt).toLocaleTimeString("de-CH", { hour: "2-digit", minute: "2-digit" })}
              </p>
            )}
          </div>
        </div>
      ) : (
        <>
          <QRCode value={slot.orderId} size={260} className="mx-auto rounded-[11px] border-2 p-1" />
          {countdown && (
            <div className="text-muted-foreground mt-3 flex items-center justify-center gap-1.5 text-xs tabular-nums">
              <Timer className="size-3.5" aria-hidden />
              Reserviert · noch {countdown}
            </div>
          )}
        </>
      )}

      {totalLines > 0 && (
        <div className="mt-8">
          <h2 className="mb-3 text-lg font-semibold">Deine Mahlzeit</h2>
          <ul className="flex flex-col gap-3">
            {slot.lines.map((l, idx) => (
              <li key={`${l.productName}-${idx}`} className="rounded-xl border p-3">
                <div className="flex items-center gap-3">
                  {l.productImage ? (
                    <div className="relative h-16 w-16 shrink-0 overflow-hidden rounded-[11px] bg-[#cec9c6]">
                      <Image
                        src={l.productImage}
                        alt={"Produktbild von " + l.productName}
                        fill
                        sizes="64px"
                        quality={90}
                        className="h-full w-full rounded-[11px] object-cover"
                      />
                    </div>
                  ) : (
                    <div className="h-16 w-16 shrink-0 rounded-[11px] bg-[#cec9c6]" aria-hidden />
                  )}
                  <div className="min-w-0 flex-1">
                    <div className={`flex justify-between gap-3 ${isRedeemed ? "items-start" : "items-center"}`}>
                      <div className="min-w-0">
                        <p className="truncate font-medium">{l.productName}</p>
                        {isRedeemed && (
                          <div className="mt-1.5 inline-flex items-center gap-1 rounded-md bg-emerald-100 px-2 py-0.5 text-xs font-medium text-emerald-800">
                            <Check className="size-3" aria-hidden />
                            Eingelöst
                          </div>
                        )}
                      </div>
                      <p className="text-muted-foreground shrink-0 text-sm tabular-nums">×{l.quantity}</p>
                    </div>
                  </div>
                </div>
              </li>
            ))}
          </ul>
        </div>
      )}

      {error && slot && (
        <div className="bg-destructive/10 text-destructive mt-6 rounded-xl p-3 text-sm" role="alert">
          {error}
        </div>
      )}

      {!isRedeemed && (
        <div className="mt-6 flex justify-center">
          <button
            type="button"
            onClick={releaseSlot}
            disabled={releasing}
            className="text-muted-foreground hover:text-destructive flex items-center gap-1.5 text-xs underline-offset-4 transition hover:underline disabled:opacity-60"
          >
            {releasing && <Loader2 className="size-3 animate-spin" aria-hidden />}
            Für jemand anderes freigeben
          </button>
        </div>
      )}

      <div ref={footerRef} className="fixed inset-x-0 bottom-0 z-50 mx-auto max-w-xl p-4">
        <Button
          asChild
          variant="outline"
          className="rounded-pill h-12 w-full bg-white text-base font-medium text-black"
        >
          <Link href={`/claim/${token}`}>
            <ArrowLeft style={{ width: 20, height: 20 }} aria-hidden />
            Zurück zur Liste
          </Link>
        </Button>
      </div>

      <style jsx>{`
        @keyframes claimPop {
          from {
            opacity: 0;
            transform: translateY(8px);
          }
          to {
            opacity: 1;
            transform: translateY(0);
          }
        }
        @keyframes checkPop {
          from {
            opacity: 0;
            transform: scale(0.4);
          }
          to {
            opacity: 1;
            transform: scale(1);
          }
        }
      `}</style>
    </div>
  )
}
