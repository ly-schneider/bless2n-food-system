"use client"

import { useRouter } from "next/navigation"
import { useEffect, useState } from "react"
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from "@/components/ui/alert-dialog"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { useAuth } from "@/contexts/auth-context"
import { API_BASE_URL } from "@/lib/api"

import type { UserRole } from "@/types"

export default function ProfilePage() {
  const { signOut, getToken, refresh } = useAuth()
  const router = useRouter()
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [email, setEmail] = useState<string>("")
  const [savingEmail, setSavingEmail] = useState(false)
  const [firstName, setFirstName] = useState<string>("")
  const [lastName, setLastName] = useState<string>("")
  const [role, setRole] = useState<UserRole | null>(null)
  const [initialEmail, setInitialEmail] = useState<string>("")
  const [emailChangeInitiated, setEmailChangeInitiated] = useState(false)
  const [verificationCode, setVerificationCode] = useState("")

  // Load current user profile via GET /v1/users (returns full user)
  useEffect(() => {
    const loadProfile = async () => {
      try {
        setError(null)
        let token = getToken()
        if (!token) {
          await refresh()
          token = getToken()
        }
        if (!token) throw new Error("Not authenticated")
        const res = await fetch(`${API_BASE_URL}/v1/users/me`, { headers: { Authorization: `Bearer ${token}` } })
        if (!res.ok) throw new Error("Failed to load profile")
        const data = (await res.json()) as { user?: { email?: string; firstName?: string; lastName?: string; role?: UserRole } }
        if (data.user?.email) { setEmail(data.user.email); setInitialEmail(data.user.email) }
        if (data.user?.firstName !== undefined) setFirstName(data.user.firstName || '')
        if (data.user?.lastName !== undefined) setLastName(data.user.lastName || '')
        if (data.user?.role) setRole(data.user.role)
        setError(null)
      } catch (e: unknown) {
        const msg = e instanceof Error ? e.message : "Failed to load profile"
        setError(msg)
      }
    }
    void loadProfile()
  }, [getToken, refresh])

  const saveProfile = async () => {
    setSavingEmail(true)
    setError(null)
    try {
      let token = getToken()
      if (!token) {
        const ok = await refresh()
        if (ok) token = getToken()
      }
      if (!token) throw new Error("Not authenticated")
      // 1) Update names (admin only)
      if (role === 'admin') {
        const res = await fetch(`${API_BASE_URL}/v1/users/me`, {
          method: "PATCH",
          headers: { "Content-Type": "application/json", Authorization: `Bearer ${token}` },
          body: JSON.stringify({ firstName, lastName }),
        })
        if (!res.ok) {
          const t = await res.text()
          throw new Error(t || "Failed to update profile")
        }
      }
      // 2) Initiate email change only if changed
      if (email && initialEmail && email.toLowerCase() !== initialEmail.toLowerCase()) {
        const res2 = await fetch(`${API_BASE_URL}/v1/users/me/email-change`, {
          method: "POST",
          headers: { "Content-Type": "application/json", Authorization: `Bearer ${token}` },
          body: JSON.stringify({ newEmail: email }),
        })
        if (!res2.ok) {
          const t = await res2.text()
          throw new Error(t || "Failed to start email change")
        }
        setEmailChangeInitiated(true)
      }
      setError(null)
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
      let token = getToken()
      if (!token) {
        const ok = await refresh()
        if (ok) token = getToken()
      }
      if (!token) throw new Error("Not authenticated")
      const res = await fetch(`${API_BASE_URL}/v1/users/me/email-change/confirm`, {
        method: "POST",
        headers: { "Content-Type": "application/json", Authorization: `Bearer ${token}` },
        body: JSON.stringify({ code: verificationCode }),
      })
      if (!res.ok) throw new Error("Ungültiger Code oder Bestätigung fehlgeschlagen")
      // Refresh auth state to get updated user
      await refresh()
      // Re-fetch profile to reflect updated email
      try {
        const token = getToken()
        if (token) {
          const r = await fetch(`${API_BASE_URL}/v1/users/me`, { headers: { Authorization: `Bearer ${token}` } })
          if (r.ok) {
            const d = (await r.json()) as { user?: { email?: string } }
            if (d.user?.email) { setEmail(d.user.email); setInitialEmail(d.user.email) }
          }
        }
      } catch {}
      setEmailChangeInitiated(false)
      setVerificationCode("")
      setError(null)
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
      let token = getToken()
      if (!token) {
        const ok = await refresh()
        if (ok) token = getToken()
      }
      if (!token) throw new Error("Not authenticated")
      const res = await fetch(`${API_BASE_URL}/v1/users/me`, {
        method: "DELETE",
        headers: { Authorization: `Bearer ${token}` },
      })
      if (!res.ok) throw new Error("Failed to delete account")
      await signOut()
      router.push("/")
      setError(null)
    } catch (e: unknown) {
      const msg = e instanceof Error ? e.message : "Failed to delete account"
      setError(msg)
    } finally {
      setDeleting(false)
    }
  }

  return (
    <div className="container mx-auto max-w-xl p-4">
      <h1 className="mb-4 text-2xl font-semibold">Profil</h1>

      {error && <p className="mb-3 text-sm text-red-600">{error}</p>}

      <section className="mt-2 mb-8 space-y-3">
        {role === 'admin' && (
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="mb-1 block text-sm font-semibold" htmlFor="firstName">Vorname</label>
              <Input className="w-full rounded-[11px] border px-3 py-2" id="firstName" value={firstName} onChange={(e) => setFirstName(e.target.value)} />
            </div>
            <div>
              <label className="mb-1 block text-sm font-semibold" htmlFor="lastName">Nachname</label>
              <Input className="w-full rounded-[11px] border px-3 py-2" id="lastName" value={lastName} onChange={(e) => setLastName(e.target.value)} />
            </div>
          </div>
        )}
        <div>
          <label className="mb-1 block text-sm font-semibold" htmlFor="email">E-Mail-Adresse</label>
          <div className="flex gap-2">
            <Input
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
              <Input className="w-full rounded-[11px] border px-3 py-2" placeholder="6-stelliger Code" value={verificationCode} onChange={(e) => setVerificationCode(e.target.value)} />
              <Button onClick={confirmEmailChange} disabled={verificationCode.length !== 6 || loading}>Bestätigen</Button>
            </div>
          </div>
        )}
      </section>

      <div className="flex justify-self-end gap-4">
        <Button
          variant={"outline"}
          className="rounded-[10px]"
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
