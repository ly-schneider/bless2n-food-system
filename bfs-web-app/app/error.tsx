"use client"

import { AlertTriangle, Home, RefreshCw } from "lucide-react"
import { useEffect } from "react"

import { Button } from "@/components/ui/button"

export default function Error({ error, reset }: { error: Error & { digest?: string }; reset: () => void }) {
  useEffect(() => {
    console.error("Application error:", error)
  }, [error])

  return (
    <div className="flex min-h-screen flex-col items-center justify-center bg-background px-4">
      <div className="mx-auto max-w-md text-center">
        <div className="mb-6 flex justify-center">
          <div className="flex size-20 items-center justify-center rounded-full bg-destructive/10">
            <AlertTriangle className="size-10 text-destructive" />
          </div>
        </div>

        <h1 className="font-primary mb-4 text-2xl sm:text-3xl">Etwas ist schiefgelaufen</h1>

        <p className="mb-8 text-muted-foreground">
          Ein unerwarteter Fehler ist aufgetreten.
        </p>

        <div className="flex flex-col gap-3 sm:flex-row sm:justify-center">
          <Button onClick={reset} className="h-10 rounded-xl px-6!">
            <RefreshCw className="size-4" />
            Erneut versuchen
          </Button>

          <Button asChild variant="outline" className="h-10 rounded-xl px-6!">
            <a href="/">
              <Home className="size-4" />
              Zur Startseite
            </a>
          </Button>
        </div>

        {error.digest && (
          <div className="mt-8 rounded-lg border border-border bg-card p-4">
            <p className="text-xs text-muted-foreground">
              Fehler-ID: <code className="rounded bg-muted px-1.5 py-0.5 font-mono">{error.digest}</code>
            </p>
            <p className="mt-2 text-xs text-muted-foreground">
              Bitte gib diese ID an, wenn du uns kontaktierst — sie hilft uns, das Problem schneller zu finden.
            </p>
          </div>
        )}
      </div>
    </div>
  )
}
