"use client"

import { useRouter, useSearchParams } from "next/navigation"
import { useEffect, useState } from "react"
import AuthHeader from "@/components/layout/auth-header"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { InputOTP, InputOTPGroup, InputOTPSlot } from "@/components/ui/input-otp"
import { Label } from "@/components/ui/label"
import { authClient, signIn, emailOtp } from "@/lib/auth/client"

type LoginStep = "email" | "otp"

export default function LoginClient() {
  const router = useRouter()
  const sp = useSearchParams()
  const next = sp.get("next") || "/"
  const { data: session, isPending } = authClient.useSession()

  const [step, setStep] = useState<LoginStep>("email")
  const [email, setEmail] = useState("")
  const [otp, setOtp] = useState("")
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  // Redirect to destination if already authenticated
  useEffect(() => {
    if (!isPending && session?.user) {
      router.replace(next)
    }
  }, [session, isPending, router, next])

  const handleGoogleSignIn = async () => {
    setLoading(true)
    setError(null)
    try {
      await signIn.social({
        provider: "google",
        callbackURL: next,
      })
    } catch (err) {
      setError("Google-Anmeldung fehlgeschlagen. Bitte erneut versuchen.")
      setLoading(false)
    }
  }

  const handleEmailSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!email) return

    setLoading(true)
    setError(null)
    try {
      const result = await emailOtp.sendVerificationOtp({
        email,
        type: "sign-in",
      })
      if (result.error) {
        setError(result.error.message || "Fehler beim Senden des Codes.")
        setLoading(false)
        return
      }
      setStep("otp")
    } catch {
      setError("Fehler beim Senden des Codes. Bitte erneut versuchen.")
    } finally {
      setLoading(false)
    }
  }

  const handleOtpSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!otp || otp.length < 6) return

    setLoading(true)
    setError(null)
    try {
      const result = await signIn.emailOtp({
        email,
        otp,
      })
      if (result.error) {
        setError(result.error.message || "Ungültiger Code. Bitte erneut versuchen.")
        setLoading(false)
        return
      }
      // Success - redirect will happen via useEffect when session updates
    } catch {
      setError("Anmeldung fehlgeschlagen. Bitte erneut versuchen.")
      setLoading(false)
    }
  }

  const handleResendOtp = async () => {
    setLoading(true)
    setError(null)
    try {
      const result = await emailOtp.sendVerificationOtp({
        email,
        type: "sign-in",
      })
      if (result.error) {
        setError(result.error.message || "Fehler beim Senden des Codes.")
      }
    } catch {
      setError("Fehler beim Senden des Codes.")
    } finally {
      setLoading(false)
    }
  }

  const handleBackToEmail = () => {
    setStep("email")
    setOtp("")
    setError(null)
  }

  return (
    <div className="flex flex-1 flex-col">
      <main className="container mx-auto px-4 pt-24 pb-10">
        <AuthHeader />
        <h1 className="mt-4 mb-2 text-center text-2xl font-semibold">Anmelden oder Registrieren</h1>

        <div className="mx-auto mt-12 max-w-sm space-y-6">
          {/* Google Sign In */}
          <Button
            variant="outline"
            className="h-11 w-full gap-2 rounded-xl"
            onClick={handleGoogleSignIn}
            disabled={loading}
          >
            <svg className="h-5 w-5" viewBox="0 0 24 24">
              <path
                fill="#4285F4"
                d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"
              />
              <path
                fill="#34A853"
                d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"
              />
              <path
                fill="#FBBC05"
                d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"
              />
              <path
                fill="#EA4335"
                d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"
              />
            </svg>
            Mit Google anmelden
          </Button>

          <div className="relative">
            <div className="absolute inset-0 flex items-center">
              <span className="w-full border-t" />
            </div>
            <div className="relative flex justify-center text-xs uppercase">
              <span className="bg-background text-muted-foreground px-2">oder</span>
            </div>
          </div>

          {/* Email OTP Flow */}
          {step === "email" ? (
            <form onSubmit={handleEmailSubmit} className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="email">E-Mail-Adresse</Label>
                <Input
                  id="email"
                  type="email"
                  placeholder="name@beispiel.ch"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  required
                  disabled={loading}
                />
              </div>
              <Button type="submit" className="h-11 w-full rounded-full" disabled={loading || !email}>
                {loading ? "Sende..." : "Code senden"}
              </Button>
            </form>
          ) : (
            <form onSubmit={handleOtpSubmit} className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="otp">Bestaetigungscode</Label>
                <p className="text-muted-foreground text-sm">Wir haben einen Code an {email} gesendet.</p>
                <InputOTP maxLength={6} value={otp} onChange={setOtp} disabled={loading} autoFocus>
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
              <Button type="submit" className="h-11 w-full rounded-full" disabled={loading || otp.length < 6}>
                {loading ? "Pruefe..." : "Anmelden"}
              </Button>
              <div className="flex justify-between text-sm">
                <button
                  type="button"
                  className="text-muted-foreground hover:text-foreground underline"
                  onClick={handleBackToEmail}
                  disabled={loading}
                >
                  Andere E-Mail
                </button>
                <button
                  type="button"
                  className="text-muted-foreground hover:text-foreground underline"
                  onClick={handleResendOtp}
                  disabled={loading}
                >
                  Code erneut senden
                </button>
              </div>
            </form>
          )}

          {error && <p className="text-center text-sm text-red-600">{error}</p>}

          <p className="text-muted-foreground mt-4 text-xs">
            Mit der Anmeldung oder Registrierung akzeptierst du automatisch unsere{" "}
            <a className="underline" href="/agb" target="_blank" rel="noopener noreferrer">
              AGB
            </a>{" "}
            und die{" "}
            <a className="underline" href="/datenschutz" target="_blank" rel="noopener noreferrer">
              Datenschutzerklaerung
            </a>
            .
          </p>
        </div>
      </main>
    </div>
  )
}
