"use client"

import { GoogleAnalytics } from "@next/third-parties/google"
import { useEffect, useState } from "react"

const COOKIE_NAME = "ga_consent"

function getCookie(name: string): string | null {
  if (typeof document === "undefined") return null
  const match = document.cookie.match(new RegExp("(?:^|; )" + name + "=([^;]*)"))
  return match && typeof match[1] === "string" ? decodeURIComponent(match[1]) : null
}

type Props = { gaId?: string }

export default function AnalyticsConsentGate({ gaId }: Props) {
  const [enabled, setEnabled] = useState(false)

  useEffect(() => {
    const sync = () => {
      const consent = getCookie(COOKIE_NAME)
      setEnabled(consent === "true")
    }
    sync()

    const onChange = (e: Event) => {
      // Update from banner without reload
      // @ts-expect-error CustomEvent detail
      const val = e?.detail?.value
      if (typeof val === "boolean") setEnabled(val)
      else sync()
    }
    window.addEventListener("ga-consent-changed", onChange as EventListener)
    return () => window.removeEventListener("ga-consent-changed", onChange as EventListener)
  }, [])

  if (!enabled || !gaId) return null
  return <GoogleAnalytics gaId={gaId} />
}
