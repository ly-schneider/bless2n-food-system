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
import { authClient } from "@/lib/auth/client"

export default function ProfilePage() {
  const { signOut, user, isLoading: authLoading } = useAuth()
  const router = useRouter()
  const [error, setError] = useState<string | null>(null)
  const [name, setName] = useState<string>("")
  const [saving, setSaving] = useState(false)

  // Initialize form with user data from Better Auth session
  useEffect(() => {
    if (user) {
      const fullName = [user.firstName, user.lastName].filter(Boolean).join(" ")
      setName(fullName)
    }
  }, [user])

  // Redirect if not authenticated
  useEffect(() => {
    if (!authLoading && !user) {
      router.push("/login?next=/profile")
    }
  }, [authLoading, user, router])

  const saveProfile = async () => {
    setSaving(true)
    setError(null)
    try {
      const result = await authClient.updateUser({ name })
      if (result.error) {
        throw new Error(result.error.message || "Profil konnte nicht aktualisiert werden")
      }
    } catch (e: unknown) {
      const msg = e instanceof Error ? e.message : "Profil konnte nicht aktualisiert werden"
      setError(msg)
    } finally {
      setSaving(false)
    }
  }

  const [deleting, setDeleting] = useState(false)
  const deleteAccount = async () => {
    setDeleting(true)
    setError(null)
    try {
      const result = await authClient.deleteUser()
      if (result.error) {
        throw new Error(result.error.message || "Konto konnte nicht gelöscht werden")
      }
      await signOut()
      router.push("/")
    } catch (e: unknown) {
      const msg = e instanceof Error ? e.message : "Konto konnte nicht gelöscht werden"
      setError(msg)
    } finally {
      setDeleting(false)
    }
  }

  if (authLoading) {
    return (
      <div className="container mx-auto max-w-xl p-4">
        <p>Lade...</p>
      </div>
    )
  }

  if (!user) {
    return null
  }

  return (
    <div className="container mx-auto max-w-xl p-4">
      <h1 className="mb-4 text-2xl font-semibold">Profil</h1>

      {error && <p className="mb-3 text-sm text-red-600">{error}</p>}

      <section className="mt-2 mb-8 space-y-3">
        <div>
          <label className="mb-1 block text-sm font-semibold" htmlFor="email">
            E-Mail-Adresse
          </label>
          <Input
            className="w-full rounded-[11px] border bg-muted px-3 py-2"
            type="email"
            value={user.email}
            disabled
            id="email"
          />
          <p className="mt-1 text-xs text-muted-foreground">
            E-Mail-Adresse kann nicht geändert werden
          </p>
        </div>

        <div>
          <label className="mb-1 block text-sm font-semibold" htmlFor="name">
            Name
          </label>
          <div className="flex gap-2">
            <Input
              className="w-full rounded-[11px] border px-3 py-2"
              value={name}
              onChange={(e) => setName(e.target.value)}
              name="name"
              id="name"
              placeholder="Dein Name"
            />
            <Button className="rounded-[10px] px-6 !py-5" onClick={saveProfile} disabled={saving}>
              {saving ? "Speichern..." : "Speichern"}
            </Button>
          </div>
        </div>
      </section>

      <div className="flex gap-4 justify-self-end">
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
            <Button variant={"link"} className="text-red-500">
              Konto löschen
            </Button>
          </AlertDialogTrigger>
          <AlertDialogContent>
            <AlertDialogHeader>
              <AlertDialogTitle>Konto wirklich löschen?</AlertDialogTitle>
              <AlertDialogDescription>
                Dieser Vorgang kann nicht rückgängig gemacht werden. Dein Konto und alle zugehörigen Daten werden
                dauerhaft gelöscht.
              </AlertDialogDescription>
            </AlertDialogHeader>
            <AlertDialogFooter>
              <AlertDialogCancel>Abbrechen</AlertDialogCancel>
              <AlertDialogAction onClick={deleteAccount} disabled={deleting}>
                {deleting ? "Löschen..." : "Ja, endgültig löschen"}
              </AlertDialogAction>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialog>
      </div>
    </div>
  )
}
