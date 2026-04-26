"use client"

import { DownloadIcon, MoreVerticalIcon, PlusSquareIcon, ShareIcon } from "lucide-react"
import { useState } from "react"
import { Button } from "@/components/ui/button"
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import { usePwaInstall } from "@/hooks/use-pwa-install"

export function InstallPrompt() {
  const { installed, ios, canPromptInstall, promptInstall } = usePwaInstall()
  const [iosInstructionsOpen, setIosInstructionsOpen] = useState(false)
  const [manualInstructionsOpen, setManualInstructionsOpen] = useState(false)

  if (installed) return null

  const handleClick = async () => {
    if (canPromptInstall) {
      const outcome = await promptInstall()
      if (outcome !== "unavailable") return
    }
    if (ios) {
      setIosInstructionsOpen(true)
      return
    }
    setManualInstructionsOpen(true)
  }

  return (
    <>
      <Button
        type="button"
        variant="outline"
        size="sm"
        onClick={handleClick}
        aria-label="Installieren"
        className="gap-1.5 px-2 sm:px-3"
      >
        <DownloadIcon className="size-4" />
        <span className="hidden sm:inline">Installieren</span>
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
      <Dialog open={manualInstructionsOpen} onOpenChange={setManualInstructionsOpen}>
        <DialogContent className="sm:max-w-sm">
          <DialogHeader>
            <DialogTitle>Zum Home-Bildschirm hinzufügen</DialogTitle>
            <DialogDescription>So installierst du den Station-Scanner über das Browser-Menü.</DialogDescription>
          </DialogHeader>
          <ol className="space-y-3 text-sm">
            <li className="flex items-start gap-3">
              <span className="bg-muted mt-0.5 inline-flex size-6 shrink-0 items-center justify-center rounded-full font-semibold">
                1
              </span>
              <span className="flex flex-wrap items-center gap-1">
                Öffne das Browser-Menü
                <MoreVerticalIcon className="inline size-4" aria-label="Menü" />
                oben rechts.
              </span>
            </li>
            <li className="flex items-start gap-3">
              <span className="bg-muted mt-0.5 inline-flex size-6 shrink-0 items-center justify-center rounded-full font-semibold">
                2
              </span>
              <span>
                Wähle <strong>App installieren</strong> oder <strong>Zur Startseite hinzufügen</strong>.
              </span>
            </li>
            <li className="flex items-start gap-3">
              <span className="bg-muted mt-0.5 inline-flex size-6 shrink-0 items-center justify-center rounded-full font-semibold">
                3
              </span>
              <span>
                Bestätige mit <strong>Installieren</strong> oder <strong>Hinzufügen</strong> und öffne den Scanner vom
                Home-Bildschirm.
              </span>
            </li>
          </ol>
        </DialogContent>
      </Dialog>
    </>
  )
}
