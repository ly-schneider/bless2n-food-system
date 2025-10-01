"use client"
import { useRouter, useSearchParams } from "next/navigation"
import { useEffect, useState } from "react"
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
  const [accepted, setAccepted] = useState(false)
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

  // Handle resend cooldown tick
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
      <main className="container mx-auto px-4 py-10">
        <h1 className="mb-6 text-2xl font-semibold">Anmelden oder Registrieren</h1>

        <div className="space-y-4">
          {step === "start" && (
            <div className="space-y-3">
              <div className="space-y-2">
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
              <div className="flex items-start gap-2 rounded-[11px] border p-3">
                <input
                  id="accept"
                  type="checkbox"
                  className="mt-1"
                  checked={accepted}
                  onChange={(e) => setAccepted(e.target.checked)}
                />
                <Label htmlFor="accept" className="flex-1 text-xs leading-5 font-normal">
                  Ich habe die <a className="underline" href="/agb" target="_blank" rel="noopener noreferrer">AGB</a> und
                  die <a className="underline" href="/datenschutz" target="_blank" rel="noopener noreferrer">Datenschutzerklärung</a> gelesen und akzeptiere sie.
                </Label>
              </div>
              <Button className="rounded-pill h-10 w-full" onClick={requestCode} disabled={loading || !email || !accepted}>
                {loading ? "Sende…" : "Code senden"}
              </Button>
              <p className="text-muted-foreground text-xs">
                Wir senden dir einen einmaligen 6‑stelligen Code per E‑Mail. Verwende nach Möglichkeit kein geteiltes
                Gerät.
              </p>
              {error && <p className="text-sm text-red-600">{error}</p>}
            </div>
          )}

          {step === "code" && (
            <div className="space-y-3">
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
