"use client"
import Image from "next/image"
import Link from "next/link"
import { useRouter, useSearchParams } from "next/navigation"
import { useEffect, useState } from "react"
import AuthHeader from "@/components/layout/auth-header"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { InputOTP, InputOTPGroup, InputOTPSlot } from "@/components/ui/input-otp"
import { Label } from "@/components/ui/label"
import { useAuth } from "@/contexts/auth-context"
import type { User } from "@/types"

export default function LoginPage() {
  const [email, setEmail] = useState("")
  const [step, setStep] = useState<"start" | "code">("start")
  const [otp, setOtp] = useState("")
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [resending, setResending] = useState(false)
  const [resendCooldown, setResendCooldown] = useState(0)
  const router = useRouter()
  const sp = useSearchParams()
  const next = sp.get("next") || "/"
  const { setAuth } = useAuth()

  const requestCode = async () => {
    setLoading(true)
    setError(null)
    try {
      await fetch("/api/auth/otp/request", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email }),
      })
      setStep("code")
      setResendCooldown(60)
    } catch {
      setError("Etwas ist schiefgelaufen. Bitte erneut versuchen.")
    } finally {
      setLoading(false)
    }
  }

  const verifyCode = async () => {
    setLoading(true)
    setError(null)
    try {
      const res = await fetch("/api/auth/otp/verify", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email, otp }),
      })
      if (!res.ok) throw new Error("Ungültiger Code")
      const data = (await res.json()) as {
        access_token: string
        expires_in: number
        user: User
        is_new?: boolean
      }
      setAuth(data.access_token, data.expires_in, data.user)
      if (data.is_new) {
        router.replace("/profile")
      } else {
        router.replace(next)
      }
    } catch {
      setError("Ungültiger Code. Bitte erneut versuchen.")
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    if (step !== "code" || resendCooldown <= 0) return
    const id = setInterval(() => {
      setResendCooldown((s) => (s > 0 ? s - 1 : 0))
    }, 1000)
    return () => clearInterval(id)
  }, [step, resendCooldown])

  const resendCode = async () => {
    if (resendCooldown > 0 || resending) return
    setResending(true)
    setError(null)
    try {
      await fetch("/api/auth/otp/request", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email }),
      })
      setResendCooldown(60)
    } catch {
      setError("Senden fehlgeschlagen. Bitte später erneut versuchen.")
    } finally {
      setResending(false)
    }
  }

  return (
    <div className="flex flex-1 flex-col">
      <main className="container mx-auto px-4 pt-24 pb-10">
        <AuthHeader />
        <h1 className="mt-4 mb-2 text-center text-2xl font-semibold">Anmelden oder Registrieren</h1>

        <div className="mt-12 space-y-4">
          {step === "start" && (
            <div className="space-y-3">
              <div className="flex flex-col gap-2">
                <Button
                  className="border-border h-10 w-full rounded-[9px] border bg-transparent"
                  variant="outline"
                  style={{ fontFamily: "Roboto, system-ui, Arial, sans-serif" }}
                  onClick={() => {
                    window.location.href = `/api/auth/google/start?next=${encodeURIComponent(next)}`
                  }}
                >
                  <span className="inline-flex items-center justify-center gap-[10px]">
                    <span
                      className="inline-flex items-center justify-center"
                      aria-hidden="true"
                      style={{ width: 18, height: 18 }}
                    >
                      <svg viewBox="0 0 48 48" width="18" height="18">
                        <path
                          fill="#FFC107"
                          d="M43.6 20.5H42V20H24v8h11.3C34.3 32.9 29.7 36 24 36c-6.6 0-12-5.4-12-12s5.4-12 12-12c3 0 5.7 1.1 7.7 3l5.7-5.7C33.9 6.1 29.2 4 24 4 12.9 4 4 12.9 4 24s8.9 20 20 20c10 0 19-7.3 19-20 0-1.3-.1-2.7-.4-3.5z"
                        />
                        <path
                          fill="#FF3D00"
                          d="M6.3 14.7l6.6 4.8C14.7 16.5 18.9 12 24 12c3 0 5.7 1.1 7.7 3l5.7-5.7C33.9 6.1 29.2 4 24 4 16 4 9.2 8.6 6.3 14.7z"
                        />
                        <path
                          fill="#4CAF50"
                          d="M24 44c5.6 0 10.3-1.9 13.7-5.2l-6.3-5.2c-1.7 1.2-4 2.1-7.4 2.1-5.7 0-10.3-3.8-12-9l-6.6 5.1C9.2 39.4 16 44 24 44z"
                        />
                        <path
                          fill="#1976D2"
                          d="M43.6 20.5H42V20H24v8h11.3c-.9 2.9-3.4 5.1-6.7 5.1-.4 0-.9 0-1.3-.1l6.3 5.2c.4.4 0 0 0 0C38.6 35.1 43 30.1 43.6 20.5z"
                        />
                      </svg>
                    </span>
                    <span className="whitespace-nowrap">Mit Google fortfahren</span>
                  </span>
                </Button>
                <div className="my-2 flex items-center gap-3">
                  <div className="bg-border h-px flex-1" />
                  <span className="text-muted-foreground text-xs">ODER</span>
                  <div className="bg-border h-px flex-1" />
                </div>
              </div>
              <div className="flex flex-col gap-2">
                <Label htmlFor="email">E-Mail-Adresse</Label>
                <Input
                  id="email"
                  type="email"
                  inputMode="email"
                  autoComplete="email"
                  placeholder="du@beispiel.ch"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                />
              </div>
              <Button className="rounded-pill h-10 w-full" onClick={requestCode} disabled={loading || !email}>
                {loading ? "Sende…" : "Code senden"}
              </Button>
              <p className="text-muted-foreground text-xs">
                Wir senden dir einen einmaligen 6‑stelligen Code per E‑Mail. Verwende nach Möglichkeit kein geteiltes
                Gerät.
              </p>
              <p className="text-muted-foreground mb-6 text-xs">
                Mit der Anmeldung oder Registrierung akzeptierst du automatisch unsere{" "}
                <a className="underline" href="/agb" target="_blank" rel="noopener noreferrer">
                  AGB
                </a>{" "}
                und die{" "}
                <a className="underline" href="/datenschutz" target="_blank" rel="noopener noreferrer">
                  Datenschutzerklärung
                </a>
                .
              </p>
              {error && <p className="text-sm text-red-600">{error}</p>}
            </div>
          )}

          {step === "code" && (
            <div className="space-y-3">
              <p className="mt-4 mb-2 text-sm">Wir haben dir einen Code per E-Mail gesendet. Bitte gib ihn hier ein:</p>

              <div className="space-y-2">
                <Label htmlFor="otp">Code eingeben</Label>
                <InputOTP
                  maxLength={6}
                  value={otp}
                  onChange={(val) => setOtp((val || "").replace(/\D+/g, ""))}
                  containerClassName="w-full mt-1"
                >
                  <InputOTPGroup>
                    <InputOTPSlot index={0} />
                    <InputOTPSlot index={1} />
                    <InputOTPSlot index={2} />
                    <InputOTPSlot index={3} />
                    <InputOTPSlot index={4} />
                    <InputOTPSlot index={5} />
                  </InputOTPGroup>
                </InputOTP>
              </div>
              <Button className="rounded-pill h-10 w-full" onClick={verifyCode} disabled={loading || otp.length < 6}>
                {loading ? "Prüfe…" : "Bestätigen"}
              </Button>
              <div className="flex items-center justify-between gap-4">
                <p className="text-muted-foreground text-xs">
                  Nicht erhalten? Spam prüfen. Wir fragen dich niemals nach deinem Code.
                </p>
                <Button variant="link" size="sm" onClick={resendCode} disabled={resendCooldown > 0 || resending}>
                  {resendCooldown > 0
                    ? `Erneut senden in ${resendCooldown}s`
                    : resending
                    ? "Sende…"
                    : "Code erneut senden"}
                </Button>
              </div>
              {error && <p className="text-sm text-red-600">{error}</p>}
            </div>
          )}
        </div>
      </main>
    </div>
  )
}
