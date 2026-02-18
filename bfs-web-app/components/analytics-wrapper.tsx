"use client"

import { usePathname } from "next/navigation"
import CookieBanner from "@/components/cookie-banner"
import AnalyticsConsentGate from "@/components/google-analytics"

const EXCLUDED_PATHS = ["/pos", "/station"]

export default function AnalyticsWrapper({ gaId }: { gaId?: string }) {
  const pathname = usePathname()

  const isExcluded = EXCLUDED_PATHS.some((path) => pathname.startsWith(path))
  if (isExcluded) {
    return null
  }

  return (
    <>
      <AnalyticsConsentGate gaId={gaId} />
      <CookieBanner />
    </>
  )
}
