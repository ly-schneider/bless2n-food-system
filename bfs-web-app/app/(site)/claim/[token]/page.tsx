"use client"

import { Loader2, LockKeyhole, Utensils } from "lucide-react"
import Image from "next/image"
import { useParams } from "next/navigation"
import { useCallback, useEffect, useRef, useState } from "react"
import QRCode from "@/components/qrcode"
import { InputOTP, InputOTPGroup, InputOTPSlot } from "@/components/ui/input-otp"
import { readErrorMessage } from "@/lib/http"
import type { ClaimCampaignResponse } from "@/types/volunteer"

const ACCESS_INPUT_PATTERN = "^[A-Za-z1-9]*$"

export default function ClaimPage() {
  const params = useParams<{ token: string }>()
  const token = params?.token

  const [authed, setAuthed] = useState<null | boolean>(null)
  const [code, setCode] = useState("")
  const [codeError, setCodeError] = useState<string | null>(null)
  const [verifying, setVerifying] = useState(false)
  const [shakeKey, setShakeKey] = useState(0)

  const [data, setData] = useState<ClaimCampaignResponse | null>(null)
  const [error, setError] = useState<string | null>(null)
  const firstLoadDone = useRef(false)

  const load = useCallback(async () => {
    if (!token) return
    try {
      const res = await fetch(`/api/v1/claim/${encodeURIComponent(token)}`, { credentials: "include" })
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
      const j = (await res.json()) as ClaimCampaignResponse
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

  const handleCodeChange = (raw: string) => {
    const next = raw.toUpperCase().replace(/[^A-Z1-9]/g, "")
    setCode(next)
    setCodeError(null)
  }

  const submitCode = useCallback(async () => {
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
        if (res.status === 401) setCodeError("Code ist falsch.")
        else if (res.status === 404) setCodeError("Diese Kampagne existiert nicht.")
        else if (res.status === 410) setCodeError("Diese Kampagne ist nicht mehr aktiv.")
        else setCodeError(await readErrorMessage(res))
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
  }, [code, load, token])

  useEffect(() => {
    if (code.length === 4 && !verifying) {
      submitCode()
    }
  }, [code, submitCode, verifying])

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

  // ── Error state ─────────────────────────────────────────────────
  if (error && !data) {
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

  if (!data) {
    return (
      <div className="mx-auto max-w-xl p-4">
        <div className="mb-2 h-8 w-40 animate-pulse rounded bg-gray-100" />
        <div className="mb-6 h-4 w-64 animate-pulse rounded bg-gray-100" />
        <div className="mx-auto h-[260px] w-[260px] animate-pulse rounded-[11px] border-2 bg-gray-100" />
      </div>
    )
  }

  return (
    <div className="mx-auto flex max-w-xl flex-col p-4 pb-8">
      <h1 className="mb-2 text-2xl font-semibold">{data.campaign.name}</h1>
      <p className="text-muted-foreground mb-6 text-sm">Zeig diesen QR-Code an der Station vor.</p>

      <QRCode value={data.qrPayload} size={260} className="mx-auto rounded-[11px] border-2 p-1" />

      {data.products.length > 0 && (
        <div className="mt-8">
          <h2 className="mb-3 text-lg font-semibold">Dein Essen</h2>
          <ul className="flex flex-col gap-3">
            {data.products.map((p) => (
              <li key={p.productId} className="rounded-xl border p-3">
                <div className="flex items-center gap-3">
                  {p.productImage ? (
                    <div className="relative h-16 w-16 shrink-0 overflow-hidden rounded-[11px] bg-[#cec9c6]">
                      <Image
                        src={p.productImage}
                        alt={"Produktbild von " + p.productName}
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
                    <div className="flex items-center justify-between gap-3">
                      <p className="truncate font-medium">{p.productName}</p>
                      <p className="text-muted-foreground shrink-0 text-sm tabular-nums">×{p.quantity}</p>
                    </div>
                  </div>
                </div>
              </li>
            ))}
          </ul>
        </div>
      )}

      {error && (
        <div className="bg-destructive/10 text-destructive mt-6 rounded-xl p-3 text-sm" role="alert">
          {error}
        </div>
      )}
    </div>
  )
}
