"use client"

import { ArrowLeft, Home, Search } from "lucide-react"
import Link from "next/link"

import { Button } from "@/components/ui/button"

export default function NotFound() {
  return (
    <div className="flex min-h-screen flex-col items-center justify-center bg-background px-4">
      <div className="mx-auto max-w-md text-center">
        <div className="mb-8">
          <h1 className="font-primary text-[120px] leading-none sm:text-[160px]">404</h1>
        </div>

        <h2 className="font-primary mb-4 text-2xl sm:text-3xl">Seite nicht gefunden</h2>

        <p className="mb-8 text-muted-foreground">
          Die Seite, die du suchst, existiert nicht oder wurde verschoben.
        </p>

        <div className="flex flex-col gap-3 sm:flex-row sm:justify-center">
          <Button asChild className="h-10 rounded-xl px-6!">
            <Link href="/food">
              <Home className="size-4" />
              Zur Speisekarte
            </Link>
          </Button>

          <Button variant="outline" className="h-10 rounded-xl px-6!" onClick={() => window.history.back()}>
            <ArrowLeft className="size-4" />
            Zurück
          </Button>
        </div>
      </div>
    </div>
  )
}
