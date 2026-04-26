"use client"

import { MaximizeIcon, MinimizeIcon } from "lucide-react"
import { useEffect, useState } from "react"
import { Button } from "@/components/ui/button"

type FullscreenDocument = Document & {
  webkitFullscreenElement?: Element | null
  webkitExitFullscreen?: () => Promise<void>
}

type FullscreenElement = HTMLElement & {
  webkitRequestFullscreen?: (options?: FullscreenOptions) => Promise<void>
}

export function FullscreenButton() {
  const [isFullscreen, setIsFullscreen] = useState(false)
  const [supported, setSupported] = useState(false)

  useEffect(() => {
    const el = document.documentElement as FullscreenElement
    const doc = document as FullscreenDocument
    setSupported(typeof el.requestFullscreen === "function" || typeof el.webkitRequestFullscreen === "function")
    const update = () => setIsFullscreen((doc.fullscreenElement ?? doc.webkitFullscreenElement) !== null)
    update()
    document.addEventListener("fullscreenchange", update)
    document.addEventListener("webkitfullscreenchange", update)
    return () => {
      document.removeEventListener("fullscreenchange", update)
      document.removeEventListener("webkitfullscreenchange", update)
    }
  }, [])

  if (!supported) return null

  const toggle = async () => {
    const el = document.documentElement as FullscreenElement
    const doc = document as FullscreenDocument
    try {
      if (doc.fullscreenElement || doc.webkitFullscreenElement) {
        if (doc.exitFullscreen) await doc.exitFullscreen()
        else if (doc.webkitExitFullscreen) await doc.webkitExitFullscreen()
        return
      }
      if (el.requestFullscreen) await el.requestFullscreen({ navigationUI: "hide" })
      else if (el.webkitRequestFullscreen) await el.webkitRequestFullscreen()
    } catch {}
  }

  return (
    <Button
      type="button"
      variant="outline"
      size="sm"
      onClick={toggle}
      aria-label={isFullscreen ? "Vollbild beenden" : "Vollbild"}
      className="gap-1.5 px-2 sm:px-3"
    >
      {isFullscreen ? <MinimizeIcon className="size-4" /> : <MaximizeIcon className="size-4" />}
      <span className="hidden sm:inline">{isFullscreen ? "Schliessen" : "Vollbild"}</span>
    </Button>
  )
}
