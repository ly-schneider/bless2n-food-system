"use client"

import { useEffect, useState } from "react"
import { Button } from "./ui/button"

const COOKIE_NAME = "ga_consent"

function getCookie(name: string): string | null {
  if (typeof document === "undefined") return null
  const match = document.cookie.match(new RegExp("(?:^|; )" + name + "=([^;]*)"))
  return match && typeof match[1] === "string" ? decodeURIComponent(match[1]) : null
}

function setCookie(name: string, value: string, days = 365) {
  if (typeof document === "undefined") return
  const maxAge = days * 24 * 60 * 60
  document.cookie = `${name}=${encodeURIComponent(value)}; path=/; max-age=${maxAge}`
}

export default function CookieBanner() {
  const [visible, setVisible] = useState(false)

  useEffect(() => {
    const consent = getCookie(COOKIE_NAME)
    setVisible(consent === null)
  }, [])

  const accept = () => {
    setCookie(COOKIE_NAME, "true")
    if (typeof window !== "undefined") {
      window.dispatchEvent(new CustomEvent("ga-consent-changed", { detail: { value: true } }))
    }
    setVisible(false)
  }

  const decline = () => {
    setCookie(COOKIE_NAME, "false")
    if (typeof window !== "undefined") {
      window.dispatchEvent(new CustomEvent("ga-consent-changed", { detail: { value: false } }))
    }
    setVisible(false)
  }

  if (!visible) return null

  return (
    <div className="fixed inset-x-0 bottom-0 z-50 p-4">
      <div className="mx-auto max-w-4xl rounded-2xl border border-gray-200 bg-white/95 p-4 shadow-lg backdrop-blur supports-backdrop-filter:bg-white/70 dark:border-gray-800 dark:bg-gray-900/90">
        <div className="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
          <p className="text-sm text-gray-700 dark:text-gray-300">
            Wir verwenden Analysetools für minimale, datenschutzfreundliche Analyse von Seiten- und Website-Besuchen.
            Keine Werbung oder seitenübergreifendes Tracking.
          </p>
          <div className="flex shrink-0 items-center gap-2">
            <Button onClick={decline} variant="outline" className="">
              Ablehnen
            </Button>
            <Button onClick={accept} variant="primary" className="rounded-[7px]">
              Akzeptieren
            </Button>
          </div>
        </div>
      </div>
    </div>
  )
}
