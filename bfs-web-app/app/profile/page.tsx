"use client"

import { useRouter } from "next/navigation"
import { useEffect, useState } from "react"
import { useAuth } from "@/contexts/auth-context"
import { API_BASE_URL } from "@/lib/api"
import { Button } from "@/components/ui/button"
import {
  AlertDialog,
  AlertDialogTrigger,
  AlertDialogContent,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogCancel,
  AlertDialogAction,
} from "@/components/ui/alert-dialog"

type Session = {
  id: string
  device: string
  ip?: string
  current?: boolean
  created_at?: string
  last_used_at?: string
}

export default function ProfilePage() {
  const { user, signOut, getToken, refresh } = useAuth()
  const router = useRouter()
  const [sessions, setSessions] = useState<Session[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [email, setEmail] = useState<string>(user?.email ?? "")
  const [savingEmail, setSavingEmail] = useState(false)
  const [firstName, setFirstName] = useState<string>(user?.firstName ?? "")
  const [lastName, setLastName] = useState<string>(user?.lastName ?? "")
  const [emailChangeInitiated, setEmailChangeInitiated] = useState(false)
  const [verificationCode, setVerificationCode] = useState("")

  // Keep email field in sync on hard refresh when user loads asynchronously
  useEffect(() => {
    if (user?.email && email === "") {
      setEmail(user.email)
    }
  }, [user?.email])

  useEffect(() => {
    const load = async () => {
      try {
        const token = getToken()
        if (!token) return
        const res = await fetch(`${API_BASE_URL}/v1/auth/sessions`, { headers: { Authorization: `Bearer ${token}` } })
        if (!res.ok) throw new Error("Failed to load sessions")
        const data = (await res.json()) as { sessions?: Session[] }
        setSessions(data.sessions ?? [])
      } catch (e: unknown) {
        const msg = e instanceof Error ? e.message : "Failed to load sessions"
        setError(msg)
      }
    }
    void load()
  }, [getToken])

  const revokeSession = async (id: string) => {
    setLoading(true)
    setError(null)
    try {
      const token = getToken()
      if (!token) throw new Error("Not authenticated")
      const res = await fetch(`${API_BASE_URL}/v1/auth/sessions/${encodeURIComponent(id)}/revoke`, {
        method: "POST",
        headers: { Authorization: `Bearer ${token}` },
      })
      if (!res.ok) throw new Error("Failed to revoke session")
      setSessions((prev) => prev.filter((s) => s.id !== id))
    } catch (e: unknown) {
      const msg = e instanceof Error ? e.message : "Failed to revoke session"
      setError(msg)
    } finally {
      setLoading(false)
    }
  }

  const saveProfile = async () => {
    setSavingEmail(true)
    setError(null)
    try {
      const token = getToken()
      if (!token) throw new Error("Not authenticated")
      const payload: Record<string, any> = { email }
      if (user?.role === 'admin') {
        payload.firstName = firstName
        payload.lastName = lastName
      }
      const res = await fetch(`${API_BASE_URL}/v1/user`, {
        method: "PUT",
        headers: { "Content-Type": "application/json", Authorization: `Bearer ${token}` },
        body: JSON.stringify(payload),
      })
      if (!res.ok) {
        const t = await res.text()
        throw new Error(t || "Failed to update profile")
      }
      const data = await res.json() as { user?: any, email_change_initiated?: boolean, message?: string }
      setEmailChangeInitiated(Boolean(data.email_change_initiated))
      // Optimistically reflect updated names/email in local state; auth context will refresh on confirm
      if (data.user?.email) setEmail(data.user.email)
      if (user?.role === 'admin') {
        if (data.user?.firstName !== undefined) setFirstName(data.user.firstName)
        if (data.user?.lastName !== undefined) setLastName(data.user.lastName)
      }
    } catch (e: unknown) {
      const msg = e instanceof Error ? e.message : "Failed to update profile"
      setError(msg)
    } finally {
      setSavingEmail(false)
    }
  }

  const confirmEmailChange = async () => {
    setLoading(true)
    setError(null)
    try {
      const token = getToken()
      if (!token) throw new Error("Not authenticated")
      const res = await fetch(`${API_BASE_URL}/v1/user/email/confirm`, {
        method: "POST",
        headers: { "Content-Type": "application/json", Authorization: `Bearer ${token}` },
        body: JSON.stringify({ code: verificationCode }),
      })
      if (!res.ok) throw new Error("Ungültiger Code oder Bestätigung fehlgeschlagen")
      // Refresh auth state to get updated user
      await refresh()
      setEmailChangeInitiated(false)
      setVerificationCode("")
    } catch (e: unknown) {
      const msg = e instanceof Error ? e.message : "Bestätigung fehlgeschlagen"
      setError(msg)
    } finally {
      setLoading(false)
    }
  }

  const [deleting, setDeleting] = useState(false)
  const deleteAccount = async () => {
    setDeleting(true)
    setError(null)
    try {
      const token = getToken()
      if (!token) throw new Error("Not authenticated")
      const res = await fetch(`${API_BASE_URL}/v1/user`, {
        method: "DELETE",
        headers: { Authorization: `Bearer ${token}` },
      })
      if (!res.ok) throw new Error("Failed to delete account")
      await signOut()
      router.push("/")
    } catch (e: unknown) {
      const msg = e instanceof Error ? e.message : "Failed to delete account"
      setError(msg)
    } finally {
      setDeleting(false)
    }
  }

  return (
    <div className="container mx-auto max-w-2xl p-4">
      <h1 className="mb-4 text-2xl font-semibold">Profil</h1>

      {error && <p className="mb-3 text-sm text-red-600">{error}</p>}

      <section className="mt-2 mb-8 space-y-3">
        {user?.role === 'admin' && (
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="mb-1 block text-sm font-semibold" htmlFor="firstName">Vorname</label>
              <input className="w-full rounded-[11px] border px-3 py-2" id="firstName" value={firstName} onChange={(e) => setFirstName(e.target.value)} />
            </div>
            <div>
              <label className="mb-1 block text-sm font-semibold" htmlFor="lastName">Nachname</label>
              <input className="w-full rounded-[11px] border px-3 py-2" id="lastName" value={lastName} onChange={(e) => setLastName(e.target.value)} />
            </div>
          </div>
        )}
        <div>
          <label className="mb-1 block text-sm font-semibold" htmlFor="email">E-Mail-Adresse</label>
          <div className="flex gap-2">
            <input
              className="w-full rounded-[11px] border px-3 py-2"
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              name="email"
              id="email"
            />
            <Button className="rounded-[10px] !py-5 px-6" onClick={saveProfile} disabled={savingEmail}>
              {savingEmail ? "Speichern…" : "Speichern"}
            </Button>
          </div>
        </div>

        {emailChangeInitiated && (
          <div className="rounded-[11px] border p-3">
            <p className="mb-2 text-sm">Wir haben dir einen Code an die neue Adresse gesendet. Bitte bestätige die Änderung:</p>
            <div className="flex gap-2">
              <input className="w-full rounded-[11px] border px-3 py-2" placeholder="6-stelliger Code" value={verificationCode} onChange={(e) => setVerificationCode(e.target.value)} />
              <Button onClick={confirmEmailChange} disabled={verificationCode.length !== 6 || loading}>Bestätigen</Button>
            </div>
          </div>
        )}
      </section>

      <div className="flex justify-self-end gap-4">
        <Button
          variant={"outline"}
          onClick={async () => {
            await signOut()
            router.push("/")
          }}
        >
          Abmelden
        </Button>
        <AlertDialog>
          <AlertDialogTrigger asChild>
            <Button variant={"link"} className="text-red-500">Account löschen</Button>
          </AlertDialogTrigger>
          <AlertDialogContent>
            <AlertDialogHeader>
              <AlertDialogTitle>Account wirklich löschen?</AlertDialogTitle>
              <AlertDialogDescription>
                Dieser Vorgang kann nicht rückgängig gemacht werden. Dein Account und alle zugehörigen Daten werden dauerhaft gelöscht.
              </AlertDialogDescription>
            </AlertDialogHeader>
            <AlertDialogFooter>
              <AlertDialogCancel>Abbrechen</AlertDialogCancel>
              <AlertDialogAction onClick={deleteAccount} disabled={deleting}>
                {deleting ? "Löschen…" : "Ja, endgültig löschen"}
              </AlertDialogAction>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialog>
      </div>
    </div>
  )
}
