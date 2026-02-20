"use client"

import * as Sentry from "@sentry/nextjs"
import { AlertTriangle, Home, RefreshCw } from "lucide-react"
import Link from "next/link"
import { useEffect } from "react"

import { Button } from "@/components/ui/button"

export default function Error({ error, reset }: { error: Error & { digest?: string }; reset: () => void }) {
  useEffect(() => {
    Sentry.captureException(error)
  }, [error])

  return (
    <div className="bg-background flex min-h-screen flex-col items-center justify-center px-4">
      <div className="mx-auto max-w-md text-center">
        <div className="mb-6 flex justify-center">
          <div className="bg-destructive/10 flex size-20 items-center justify-center rounded-full">
            <AlertTriangle className="text-destructive size-10" />
          </div>
        </div>

        <h1 className="font-primary mb-4 text-2xl sm:text-3xl">Etwas ist schiefgelaufen</h1>

        <p className="text-muted-foreground mb-8">Ein unerwarteter Fehler ist aufgetreten.</p>

        <div className="flex flex-col gap-3 sm:flex-row sm:justify-center">
          <Button onClick={reset} className="h-10 rounded-xl px-6!">
            <RefreshCw className="size-4" />
            Erneut versuchen
          </Button>

          <Button asChild variant="outline" className="h-10 rounded-xl px-6!">
            <Link href="/">
              <Home className="size-4" />
              Zur Startseite
            </Link>
          </Button>
        </div>

        {error.digest && (
          <div className="border-border bg-card mt-8 rounded-lg border p-4">
            <p className="text-muted-foreground text-xs">
              Fehler-ID: <code className="bg-muted rounded px-1.5 py-0.5 font-mono">{error.digest}</code>
            </p>
            <p className="text-muted-foreground mt-2 text-xs">
              Bitte gib diese ID an, wenn du uns kontaktierst — sie hilft uns, das Problem schneller zu finden.
            </p>
          </div>
        )}
      </div>
    </div>
  )
}
