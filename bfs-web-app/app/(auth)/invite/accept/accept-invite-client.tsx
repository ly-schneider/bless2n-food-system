"use client"

import { useRouter, useSearchParams } from "next/navigation"
import { useEffect, useMemo, useState } from "react"
import AuthHeader from "@/components/layout/auth-header"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { authClient } from "@/lib/auth-client"

export default function AcceptInviteClient() {
  const sp = useSearchParams()
  const token = sp.get("token") || ""
  const [loading, setLoading] = useState(true)
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [success, setSuccess] = useState(false)
  const [email, setEmail] = useState("")
  const [firstName, setFirstName] = useState("")
  const [lastName, setLastName] = useState("")
  const router = useRouter()
  const { data: session } = authClient.useSession()

  const disabled = useMemo(() => submitting || !firstName || !token, [submitting, firstName, token])

  // If user is already authenticated, redirect to admin
  useEffect(() => {
    if (session?.user && success) {
      router.replace("/admin")
    }
  }, [session, success, router])

  useEffect(() => {
    let cancelled = false
    async function verify() {
      if (!token) {
        setError("Ungültiger Link. Token fehlt.")
        setLoading(false)
        return
      }
      try {
        const res = await fetch(`/api/v1/invites/verify`, {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ token }),
        })
        if (!res.ok) {
          setError("Diese Einladung ist ungültig oder abgelaufen.")
          setLoading(false)
          return
        }
        const data = (await res.json()) as { inviteeEmail: string; status: string }
        if (!cancelled) {
          setEmail(data.inviteeEmail)
          setLoading(false)
        }
      } catch {
        if (!cancelled) {
          setError("Verifizierung fehlgeschlagen. Bitte spaeter erneut versuchen.")
          setLoading(false)
        }
      }
    }
    verify()
    return () => {
      cancelled = true
    }
  }, [token])

  const acceptInvite = async () => {
    setSubmitting(true)
    setError(null)
    try {
      const res = await fetch("/api/v1/invites/accept", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ token, firstName, lastName: lastName || null }),
      })
      if (!res.ok) {
        const msg = res.status === 401 ? "Einladung abgelaufen oder bereits verwendet." : "Annahme fehlgeschlagen."
        setError(msg)
        setSubmitting(false)
        return
      }
      // Invite accepted successfully
      setSuccess(true)
    } catch {
      setError("Etwas ist schiefgelaufen. Bitte erneut versuchen.")
      setSubmitting(false)
    }
  }

  if (success) {
    return (
      <div className="flex flex-1 flex-col">
        <main className="container mx-auto px-4 pt-24 pb-10">
          <AuthHeader />
          <h1 className="mt-4 mb-2 text-center text-2xl font-semibold">Einladung angenommen</h1>
          <div className="mt-12 space-y-4 text-center">
            <p>Deine Admin-Einladung wurde erfolgreich angenommen.</p>
            <p className="text-muted-foreground text-sm">
              Melde dich jetzt an, um auf das Admin-Dashboard zuzugreifen.
            </p>
            <Button className="rounded-pill h-10" onClick={() => router.push("/login?next=/admin")}>
              Zur Anmeldung
            </Button>
          </div>
        </main>
      </div>
    )
  }

  return (
    <div className="flex flex-1 flex-col">
      <main className="container mx-auto px-4 pt-24 pb-10">
        <AuthHeader />
        <h1 className="mt-4 mb-2 text-center text-2xl font-semibold">Admin-Einladung annehmen</h1>

        <div className="mt-12 space-y-4">
          {loading ? (
            <p>Pruefe Einladung...</p>
          ) : (
            <div className="space-y-3">
              <div className="flex flex-col gap-2">
                <Label htmlFor="email">E-Mail</Label>
                <Input id="email" value={email} disabled className="bg-muted" />
              </div>
              <div className="flex flex-col gap-2">
                <Label htmlFor="firstName">Vorname</Label>
                <Input
                  id="firstName"
                  value={firstName}
                  onChange={(e) => setFirstName(e.target.value)}
                  placeholder="Max"
                />
              </div>
              <div className="flex flex-col gap-2">
                <Label htmlFor="lastName">Nachname</Label>
                <Input
                  id="lastName"
                  value={lastName}
                  onChange={(e) => setLastName(e.target.value)}
                  placeholder="Mustermann"
                />
              </div>
              <Button className="rounded-pill h-10 w-full" onClick={acceptInvite} disabled={disabled}>
                {submitting ? "Sende..." : "Einladung annehmen"}
              </Button>
              {error && <p className="text-sm text-red-600">{error}</p>}
            </div>
          )}
        </div>
      </main>
    </div>
  )
}
