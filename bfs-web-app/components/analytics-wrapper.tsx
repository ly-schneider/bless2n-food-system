"use client"

import Script from "next/script"
import { usePathname } from "next/navigation"

const EXCLUDED_PATHS = ["/pos", "/station"]

export default function AnalyticsWrapper() {
  const pathname = usePathname()
  const websiteId = process.env.NEXT_PUBLIC_UMAMI_WEBSITE_ID

  if (!websiteId || EXCLUDED_PATHS.some((path) => pathname.startsWith(path))) {
    return null
  }

  return <Script defer src="https://cloud.umami.is/script.js" data-website-id={websiteId} />
}
