"use client"

import { DownloadIcon, MaximizeIcon, PlusSquareIcon, ShareIcon, ZapIcon } from "lucide-react"
import { useEffect, useState } from "react"
import { Button } from "@/components/ui/button"
import {
  Drawer,
  DrawerClose,
  DrawerContent,
  DrawerDescription,
  DrawerFooter,
  DrawerHeader,
  DrawerTitle,
} from "@/components/ui/drawer"
import { usePwaInstall } from "@/hooks/use-pwa-install"

const DISMISS_KEY = "bfs.stationInstallHintDismissedAt"
const DISMISS_WINDOW_MS = 7 * 24 * 60 * 60 * 1000

function wasRecentlyDismissed(): boolean {
  if (typeof window === "undefined") return true
  try {
    const raw = window.localStorage.getItem(DISMISS_KEY)
    if (!raw) return false
    const ts = Number.parseInt(raw, 10)
    if (!Number.isFinite(ts)) return false
    return Date.now() - ts < DISMISS_WINDOW_MS
  } catch {
    return false
  }
}

function rememberDismissal(): void {
  if (typeof window === "undefined") return
  try {
    window.localStorage.setItem(DISMISS_KEY, String(Date.now()))
  } catch {}
}

export function InstallHintDrawer() {
  const { installed, ios, canPromptInstall, canShowIosInstructions, promptInstall } = usePwaInstall()
  const [open, setOpen] = useState(false)

  const canOffer = canPromptInstall || canShowIosInstructions

  useEffect(() => {
    if (installed) {
      setOpen(false)
      return
    }
    if (!canOffer) return
    if (wasRecentlyDismissed()) return
    const t = window.setTimeout(() => setOpen(true), 600)
    return () => window.clearTimeout(t)
  }, [installed, canOffer])

  const handleOpenChange = (next: boolean) => {
    setOpen(next)
    if (!next) rememberDismissal()
  }

  const handleInstall = async () => {
    const outcome = await promptInstall()
    if (outcome === "accepted") {
      setOpen(false)
      return
    }
    if (outcome === "dismissed") {
      setOpen(false)
      rememberDismissal()
    }
  }

  if (installed || !canOffer) return null

  return (
    <Drawer open={open} onOpenChange={handleOpenChange}>
      <DrawerContent className="mx-auto w-full sm:max-w-md">
        <DrawerHeader className="items-start gap-2 pt-6">
          <div className="bg-primary/10 text-primary mb-3 flex size-12 items-center justify-center rounded-2xl">
            <DownloadIcon className="size-6" />
          </div>
          <DrawerTitle className="text-xl">Station-Scanner installieren</DrawerTitle>
          <DrawerDescription>
            Starte den Scanner direkt vom Home-Bildschirm — ohne Browser-Leisten, im Vollbild und mit einem Tap bereit.
          </DrawerDescription>
        </DrawerHeader>

        <ul className="space-y-2 px-5 pb-2 text-sm">
          <li className="flex items-center gap-3">
            <MaximizeIcon className="text-muted-foreground size-4" />
            <span>Vollbildmodus ohne Ablenkung</span>
          </li>
          <li className="flex items-center gap-3">
            <ZapIcon className="text-muted-foreground size-4" />
            <span>Schnellerer Start direkt ins Scanner-Fenster</span>
          </li>
          <li className="flex items-center gap-3">
            <DownloadIcon className="text-muted-foreground size-4" />
            <span>Funktioniert wie eine native App</span>
          </li>
        </ul>

        {ios ? (
          <ol className="bg-muted/40 mx-5 my-4 space-y-3 rounded-lg border p-4 text-sm">
            <li className="flex items-start gap-3">
              <span className="bg-background mt-0.5 inline-flex size-6 shrink-0 items-center justify-center rounded-full font-semibold">
                1
              </span>
              <span className="flex flex-wrap items-center gap-1">
                Tippe unten in Safari auf
                <ShareIcon className="inline size-4" aria-label="Teilen" />
                <strong>Teilen</strong>.
              </span>
            </li>
            <li className="flex items-start gap-3">
              <span className="bg-background mt-0.5 inline-flex size-6 shrink-0 items-center justify-center rounded-full font-semibold">
                2
              </span>
              <span className="flex flex-wrap items-center gap-1">
                Wähle
                <PlusSquareIcon className="inline size-4" aria-label="Plus" />
                <strong>Zum Home-Bildschirm</strong>.
              </span>
            </li>
            <li className="flex items-start gap-3">
              <span className="bg-background mt-0.5 inline-flex size-6 shrink-0 items-center justify-center rounded-full font-semibold">
                3
              </span>
              <span>Öffne den Station-Scanner dann vom Home-Bildschirm.</span>
            </li>
          </ol>
        ) : null}

        <DrawerFooter className="gap-2 pt-2">
          {canPromptInstall ? (
            <Button type="button" variant="primary" size="lg" onClick={handleInstall} className="gap-2">
              <DownloadIcon className="size-4" />
              Jetzt installieren
            </Button>
          ) : null}
          <DrawerClose asChild>
            <Button type="button" variant="ghost" size="lg">
              Später
            </Button>
          </DrawerClose>
        </DrawerFooter>
      </DrawerContent>
    </Drawer>
  )
}
