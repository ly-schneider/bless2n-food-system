"use client"
import { useRouter, useSearchParams } from "next/navigation"
import { useEffect, useMemo, useState } from "react"
import AuthHeader from "@/components/layout/auth-header"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { useAuth } from "@/contexts/auth-context"

import type { User } from "@/types"

export default function AcceptInviteClient() {
  const sp = useSearchParams()
  const token = sp.get("token") || ""
  const [loading, setLoading] = useState(true)
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [email, setEmail] = useState("")
  const [firstName, setFirstName] = useState("")
  const [lastName, setLastName] = useState("")
  const { setAuth } = useAuth()
  const router = useRouter()

  const disabled = useMemo(() => submitting || !firstName || !token, [submitting, firstName, token])

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
        const data = (await res.json()) as { email: string; status: string }
        if (!cancelled) {
          setEmail(data.email)
          setLoading(false)
        }
      } catch {
        if (!cancelled) {
          setError("Verifizierung fehlgeschlagen. Bitte später erneut versuchen.")
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
      const res = await fetch("/api/invites/accept", {
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
      const data = (await res.json()) as { access_token: string; expires_in: number; user: User }
      setAuth(data.access_token, data.expires_in, data.user)
      router.replace("/admin")
    } catch {
      setError("Etwas ist schiefgelaufen. Bitte erneut versuchen.")
      setSubmitting(false)
    }
  }

  return (
    <div className="flex flex-1 flex-col">
      <main className="container mx-auto px-4 pt-24 pb-10">
        <AuthHeader />
        <h1 className="mt-4 mb-2 text-center text-2xl font-semibold">Admin-Einladung annehmen</h1>

        <div className="mt-12 space-y-4">
          {loading ? (
            <p>Prüfe Einladung…</p>
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
                {submitting ? "Sende…" : "Einladung annehmen"}
              </Button>
              {error && <p className="text-sm text-red-600">{error}</p>}
            </div>
          )}
        </div>
      </main>
    </div>
  )
}
