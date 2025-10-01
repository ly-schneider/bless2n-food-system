"use client"
import { useRouter } from 'next/navigation'
import { useEffect, useState } from 'react'
import { useAuth } from '@/contexts/auth-context'

export function AuthNudgeBanner() {
  const { accessToken } = useAuth()
  const [dismissed, setDismissed] = useState(false)
  const router = useRouter()

  useEffect(() => {
    if (typeof sessionStorage !== 'undefined') {
      setDismissed(sessionStorage.getItem('auth_nudge_dismissed') === '1')
    }
  }, [])

  if (accessToken || dismissed) return null

  return (
    <div className="mb-3 rounded-lg border bg-white p-3 shadow-sm">
      <p className="text-sm">
        Gastbestellung: <span className="font-medium">Ohne Konto wird die Bestellhistorie nur lokal auf diesem Gerät gespeichert</span> und ggf. keine E‑Mail versendet (E‑Mail‑Eingabe bei TWINT ist optional).
      </p>
      <p className="text-muted-foreground mt-1 text-xs">
        Mit Anmeldung werden Bestellungen im Konto gespeichert und per E‑Mail zugestellt.
      </p>
      <div className="mt-2 flex gap-2">
        <button
          className="rounded-md bg-black px-3 py-2 text-sm text-white"
          onClick={() => router.push('/login?next=/checkout')}
        >
          Anmelden
        </button>
        <button
          className="rounded-md border px-3 py-2 text-sm"
          onClick={() => {
            if (typeof sessionStorage !== 'undefined') sessionStorage.setItem('auth_nudge_dismissed', '1')
            setDismissed(true)
          }}
        >
          Später
        </button>
      </div>
    </div>
  )
}
