"use client"

import { DownloadIcon, PlusSquareIcon, ShareIcon } from "lucide-react"
import { useEffect, useState } from "react"
import { Button } from "@/components/ui/button"
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from "@/components/ui/dialog"

type BeforeInstallPromptEvent = Event & {
  prompt: () => Promise<void>
  userChoice: Promise<{ outcome: "accepted" | "dismissed" }>
}

function isStandalone(): boolean {
  if (typeof window === "undefined") return false
  if (window.matchMedia("(display-mode: standalone)").matches) return true
  if (window.matchMedia("(display-mode: fullscreen)").matches) return true
  const nav = window.navigator as Navigator & { standalone?: boolean }
  return nav.standalone === true
}

function isIos(): boolean {
  if (typeof window === "undefined") return false
  const ua = window.navigator.userAgent
  const iOS = /iPad|iPhone|iPod/.test(ua)
  const iPadOS = ua.includes("Mac") && "ontouchend" in document
  return iOS || iPadOS
}

export function InstallPrompt() {
  const [deferred, setDeferred] = useState<BeforeInstallPromptEvent | null>(null)
  const [installed, setInstalled] = useState(false)
  const [iosInstructionsOpen, setIosInstructionsOpen] = useState(false)
  const [ios, setIos] = useState(false)

  useEffect(() => {
    setInstalled(isStandalone())
    setIos(isIos())

    const onBeforeInstall = (e: Event) => {
      e.preventDefault()
      setDeferred(e as BeforeInstallPromptEvent)
    }
    const onInstalled = () => {
      setInstalled(true)
      setDeferred(null)
    }

    window.addEventListener("beforeinstallprompt", onBeforeInstall)
    window.addEventListener("appinstalled", onInstalled)
    return () => {
      window.removeEventListener("beforeinstallprompt", onBeforeInstall)
      window.removeEventListener("appinstalled", onInstalled)
    }
  }, [])

  if (installed) return null

  const handleClick = async () => {
    if (deferred) {
      await deferred.prompt()
      const choice = await deferred.userChoice
      if (choice.outcome === "accepted") setInstalled(true)
      setDeferred(null)
      return
    }
    if (ios) {
      setIosInstructionsOpen(true)
      return
    }
  }

  const canShow = deferred !== null || ios
  if (!canShow) return null

  return (
    <>
      <Button type="button" variant="outline" size="sm" onClick={handleClick} className="gap-2">
        <DownloadIcon className="size-4" />
        Installieren
      </Button>
      <Dialog open={iosInstructionsOpen} onOpenChange={setIosInstructionsOpen}>
        <DialogContent className="sm:max-w-sm">
          <DialogHeader>
            <DialogTitle>Zum Home-Bildschirm hinzufügen</DialogTitle>
            <DialogDescription>So installierst du den Station-Scanner auf deinem iPhone oder iPad.</DialogDescription>
          </DialogHeader>
          <ol className="space-y-3 text-sm">
            <li className="flex items-start gap-3">
              <span className="bg-muted mt-0.5 inline-flex size-6 shrink-0 items-center justify-center rounded-full font-semibold">
                1
              </span>
              <span className="flex flex-wrap items-center gap-1">
                Tippe unten in Safari auf
                <ShareIcon className="inline size-4" aria-label="Teilen" />
                <strong>Teilen</strong>.
              </span>
            </li>
            <li className="flex items-start gap-3">
              <span className="bg-muted mt-0.5 inline-flex size-6 shrink-0 items-center justify-center rounded-full font-semibold">
                2
              </span>
              <span className="flex flex-wrap items-center gap-1">
                Wähle
                <PlusSquareIcon className="inline size-4" aria-label="Plus" />
                <strong>Zum Home-Bildschirm</strong>.
              </span>
            </li>
            <li className="flex items-start gap-3">
              <span className="bg-muted mt-0.5 inline-flex size-6 shrink-0 items-center justify-center rounded-full font-semibold">
                3
              </span>
              <span>
                Bestätige mit <strong>Hinzufügen</strong>. Öffne dann den Station-Scanner vom Home-Bildschirm.
              </span>
            </li>
          </ol>
        </DialogContent>
      </Dialog>
    </>
  )
}
