"use client"

import { ArrowRight, Check, Loader2, LockKeyhole, QrCode, Utensils } from "lucide-react"
import { useParams, useRouter } from "next/navigation"
import { useCallback, useEffect, useRef, useState } from "react"
import { InputOTP, InputOTPGroup, InputOTPSlot } from "@/components/ui/input-otp"
import { readErrorMessage } from "@/lib/http"
import type { ClaimListResponse, ClaimSlotSummary } from "@/types/volunteer"

const ACCESS_INPUT_PATTERN = "^[A-Za-z1-9]*$"

export default function ClaimListPage() {
  const params = useParams<{ token: string }>()
  const router = useRouter()
  const token = params?.token

  const [authed, setAuthed] = useState<null | boolean>(null)
  const [code, setCode] = useState("")
  const [codeError, setCodeError] = useState<string | null>(null)
  const [verifying, setVerifying] = useState(false)
  const [shakeKey, setShakeKey] = useState(0)

  const [data, setData] = useState<ClaimListResponse | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [actingOnSlotId, setActingOnSlotId] = useState<string | null>(null)
  const firstLoadDone = useRef(false)

  const load = useCallback(async () => {
    if (!token) return
    try {
      const res = await fetch(`/api/v1/claim/${encodeURIComponent(token)}/slots`, { credentials: "include" })
      if (res.status === 401) {
        setAuthed(false)
        return
      }
      if (res.status === 404) {
        setAuthed(true)
        setError("Diese Kampagne existiert nicht.")
        return
      }
      if (res.status === 410) {
        setAuthed(true)
        setError("Diese Kampagne ist nicht mehr aktiv.")
        return
      }
      if (!res.ok) throw new Error(await readErrorMessage(res))
      const j = (await res.json()) as ClaimListResponse
      setAuthed(true)
      setError(null)
      setData(j)
    } catch (e: unknown) {
      setAuthed(true)
      setError(e instanceof Error ? e.message : "Laden fehlgeschlagen")
    } finally {
      firstLoadDone.current = true
    }
  }, [token])

  useEffect(() => {
    load()
  }, [load])

  useEffect(() => {
    if (authed !== true) return
    const t = window.setInterval(load, 5000)
    return () => window.clearInterval(t)
  }, [authed, load])

  const handleCodeChange = (raw: string) => {
    const next = raw.toUpperCase().replace(/[^A-Z1-9]/g, "")
    setCode(next)
    setCodeError(null)
  }

  async function submitCode() {
    if (!token) return
    setCodeError(null)
    if (code.length !== 4) {
      setCodeError("Bitte den 4-stelligen Code eingeben.")
      setShakeKey((k) => k + 1)
      return
    }
    setVerifying(true)
    try {
      const res = await fetch(`/api/v1/claim/${encodeURIComponent(token)}/auth`, {
        method: "POST",
        credentials: "include",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ code }),
      })
      if (!res.ok) {
        if (res.status === 401) {
          setCodeError("Code ist falsch.")
        } else if (res.status === 404) {
          setCodeError("Diese Kampagne existiert nicht.")
        } else if (res.status === 410) {
          setCodeError("Diese Kampagne ist nicht mehr aktiv.")
        } else {
          setCodeError(await readErrorMessage(res))
        }
        setShakeKey((k) => k + 1)
        setCode("")
        return
      }
      await load()
    } catch (e: unknown) {
      setCodeError(e instanceof Error ? e.message : "Prüfung fehlgeschlagen")
    } finally {
      setVerifying(false)
    }
  }

  useEffect(() => {
    if (code.length === 4 && !verifying) {
      submitCode()
    }
  }, [code])

  async function reserveAndOpen(slotId: string) {
    if (!token) return
    setActingOnSlotId(slotId)
    try {
      const res = await fetch(
        `/api/v1/claim/${encodeURIComponent(token)}/slots/${encodeURIComponent(slotId)}/reserve`,
        {
          method: "POST",
          credentials: "include",
        }
      )
      if (!res.ok && res.status !== 204) {
        setError(await readErrorMessage(res))
        await load()
        return
      }
      router.push(`/claim/${token}/${slotId}`)
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : "Reservieren fehlgeschlagen")
    } finally {
      setActingOnSlotId(null)
    }
  }

  function openReserved(slotId: string) {
    router.push(`/claim/${token}/${slotId}`)
  }

  if (!token) return null

  // ── Code gate ─────────────────────────────────────────────────────
  if (authed === false || (authed === null && !firstLoadDone.current)) {
    return (
      <div className="mx-auto max-w-xl p-4">
        <div className="mx-auto flex min-h-[calc(100vh-10rem)] max-w-sm flex-col justify-center gap-8 pb-16">
          <div className="flex flex-col items-center gap-4 text-center">
            <div className="bg-primary/10 text-primary flex h-14 w-14 items-center justify-center rounded-full">
              <LockKeyhole className="size-6" aria-hidden />
            </div>
            <div className="flex flex-col gap-2">
              <h1 className="text-2xl font-semibold">Mitarbeiter-Essen</h1>
              <p className="text-muted-foreground text-sm">Gib den 4-stelligen Code ein, um deine Mahlzeit zu holen.</p>
            </div>
          </div>

          <form
            onSubmit={(e) => {
              e.preventDefault()
              submitCode()
            }}
            className="flex flex-col items-center gap-4"
          >
            <div key={shakeKey} className={codeError ? "animate-[shake_220ms_ease-in-out]" : undefined}>
              <InputOTP
                maxLength={4}
                value={code}
                onChange={handleCodeChange}
                pattern={ACCESS_INPUT_PATTERN}
                inputMode="text"
                autoFocus
                disabled={verifying}
                aria-invalid={!!codeError}
                aria-describedby={codeError ? "code-error" : undefined}
              >
                <InputOTPGroup className="gap-2">
                  <InputOTPSlot index={0} className="h-16 w-14 rounded-[11px] text-2xl font-semibold uppercase" />
                  <InputOTPSlot index={1} className="h-16 w-14 rounded-[11px] text-2xl font-semibold uppercase" />
                  <InputOTPSlot index={2} className="h-16 w-14 rounded-[11px] text-2xl font-semibold uppercase" />
                  <InputOTPSlot index={3} className="h-16 w-14 rounded-[11px] text-2xl font-semibold uppercase" />
                </InputOTPGroup>
              </InputOTP>
            </div>
            <div className="flex h-5 items-center justify-center">
              {codeError ? (
                <p id="code-error" className="text-destructive text-sm" role="alert">
                  {codeError}
                </p>
              ) : verifying ? (
                <p className="text-muted-foreground flex items-center gap-2 text-sm">
                  <Loader2 className="size-4 animate-spin" aria-hidden />
                  Prüfe …
                </p>
              ) : null}
            </div>
          </form>
        </div>

        <style jsx>{`
          @keyframes shake {
            0%,
            100% {
              transform: translateX(0);
            }
            20% {
              transform: translateX(-6px);
            }
            40% {
              transform: translateX(6px);
            }
            60% {
              transform: translateX(-4px);
            }
            80% {
              transform: translateX(4px);
            }
          }
        `}</style>
      </div>
    )
  }

  // ── Error states ──────────────────────────────────────────────────
  if (error && (!data || (data.available.length === 0 && data.reservedByMe.length === 0))) {
    return (
      <div className="mx-auto max-w-xl p-4">
        <h1 className="mb-2 text-2xl font-semibold">Mitarbeiter-Essen</h1>
        <div className="mt-8 flex flex-col items-center gap-3 text-center">
          <Utensils className="text-muted-foreground size-9" aria-hidden />
          <p className="text-destructive">{error}</p>
        </div>
      </div>
    )
  }

  // ── Loading state ────────────────────────────────────────────────
  if (!data) {
    return (
      <div className="mx-auto max-w-xl p-4">
        <h1 className="mb-2 text-2xl font-semibold">Mitarbeiter-Essen</h1>
        <p className="text-muted-foreground mb-6 text-sm">Lade …</p>
        <ul className="flex flex-col gap-3">
          {[0, 1, 2].map((i) => (
            <li key={i} className="h-[68px] animate-pulse rounded-xl border bg-gray-50" />
          ))}
        </ul>
      </div>
    )
  }

  const totalConsumed = data.totalSlots - data.availableCount
  const progressPct = data.totalSlots > 0 ? Math.round((totalConsumed / data.totalSlots) * 100) : 0

  return (
    <div className="mx-auto max-w-xl p-4 pb-8">
      <h1 className="mb-2 text-2xl font-semibold">{data.campaign.name}</h1>
      <p className="text-muted-foreground text-sm tabular-nums">
        <span className="text-foreground font-medium">{data.availableCount}</span> von {data.totalSlots} verfügbar
      </p>

      <div className="bg-muted relative mt-3 mb-6 h-1.5 overflow-hidden rounded-full">
        <div
          className="bg-primary absolute inset-y-0 left-0 rounded-full transition-[width] duration-300"
          style={{ width: `${progressPct}%` }}
        />
      </div>

      {error && <div className="bg-destructive/10 text-destructive mb-4 rounded-xl p-3 text-sm">{error}</div>}

      {data.reservedByMe.length > 0 && (
        <section className="mb-6">
          <h2 className="mb-3 text-lg font-semibold">Für dich reserviert</h2>
          <ul className="flex flex-col gap-3">
            {data.reservedByMe.map((s) => (
              <li key={s.id}>
                <ClaimRow slot={s} kind="mine" onClick={() => openReserved(s.id)} loading={actingOnSlotId === s.id} />
              </li>
            ))}
          </ul>
        </section>
      )}

      <section>
        <h2 className="mb-3 text-lg font-semibold">
          {data.reservedByMe.length > 0 ? "Weitere verfügbar" : "Verfügbar"}
        </h2>
        {data.available.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-12 text-center">
            <Utensils className="text-muted-foreground mb-4 size-9" aria-hidden />
            <p className="text-muted-foreground text-lg font-semibold">Alle QR-Codes vergeben</p>
            <p className="text-muted-foreground mt-1 text-sm">
              Alle Mahlzeiten sind reserviert oder bereits eingelöst.
            </p>
          </div>
        ) : (
          <ul className="flex flex-col gap-3">
            {data.available.map((s) => (
              <li key={s.id}>
                <ClaimRow
                  slot={s}
                  kind="available"
                  onClick={() => reserveAndOpen(s.id)}
                  loading={actingOnSlotId === s.id}
                />
              </li>
            ))}
          </ul>
        )}
      </section>
    </div>
  )
}

function ClaimRow({
  kind,
  onClick,
  loading,
}: {
  slot: ClaimSlotSummary
  kind: "available" | "mine"
  onClick: () => void
  loading: boolean
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      disabled={loading}
      className={[
        "group flex min-h-[64px] w-full items-center gap-4 rounded-xl border p-3 text-left transition",
        "focus-visible:ring-ring/50 focus-visible:ring-2 focus-visible:outline-none",
        "active:scale-[0.995] disabled:cursor-not-allowed disabled:opacity-60",
        kind === "mine" ? "border-primary/40 bg-primary/5" : "bg-background hover:bg-accent/30",
      ].join(" ")}
    >
      <span
        className={[
          "shrink-0 rounded-[10px] border p-2",
          kind === "mine" ? "border-primary/40 bg-background text-primary" : "border-border bg-background",
        ].join(" ")}
      >
        {kind === "mine" ? <Check className="size-6" aria-hidden /> : <QrCode className="size-6" aria-hidden />}
      </span>
      <span className="min-w-0 flex-1 text-base font-medium">Mahlzeit</span>
      <span className="border-border bg-background shrink-0 rounded-[7px] border p-2">
        {loading ? (
          <Loader2 className="size-4 animate-spin" aria-hidden />
        ) : (
          <ArrowRight className="size-4 transition-transform group-hover:translate-x-0.5" aria-hidden />
        )}
      </span>
    </button>
  )
}
