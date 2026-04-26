"use client"

import { usePathname } from "next/navigation"
import Script from "next/script"

const UMAMI_WEBSITE_ID = "e910e40b-977e-459e-9f12-6cb95f44fc67"
const EXCLUDED_PATHS = ["/pos", "/station"]

export default function AnalyticsWrapper() {
  const pathname = usePathname()

  if (process.env.NODE_ENV !== "production" || EXCLUDED_PATHS.some((path) => pathname.startsWith(path))) {
    return null
  }

  return <Script defer src="https://cloud.umami.is/script.js" data-website-id={UMAMI_WEBSITE_ID} />
}
