"use client"
import { useRouter } from "next/navigation"
import { useEffect, useState } from "react"
import { useAuth } from "@/contexts/auth-context"
import { Button } from "../ui/button"

export function AuthNudgeBanner() {
  const { user } = useAuth()
  const [dismissed, setDismissed] = useState(false)
  const router = useRouter()

  useEffect(() => {
    if (typeof sessionStorage !== "undefined") {
      setDismissed(sessionStorage.getItem("auth_nudge_dismissed") === "1")
    }
  }, [])

  // Show nudge if user is not authenticated
  if (user || dismissed) return null

  return (
    <div className="mb-3 rounded-lg border bg-white p-3 shadow-sm">
      <p className="text-sm">Melde dich an, um deine Bestellungen zu speichern und den Checkout zu beschleunigen.</p>
      <div className="mt-4 flex gap-2">
        <Button variant="selected" onClick={() => router.push("/login?next=/food/checkout")}>
          Anmelden
        </Button>
        <Button
          variant="outline"
          onClick={() => {
            if (typeof sessionStorage !== "undefined") sessionStorage.setItem("auth_nudge_dismissed", "1")
            setDismissed(true)
          }}
        >
          Später
        </Button>
      </div>
    </div>
  )
}
