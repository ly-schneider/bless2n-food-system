import type { Metadata } from "next"
import { Golos_Text } from "next/font/google"
import AuthHeader from "@/components/layout/auth-header"
import AuthFooter from "@/components/layout/auth-footer"
import AnalyticsConsentGate from "@/components/google-analytics"
import CookieBanner from "@/components/cookie-banner"

const golosText = Golos_Text({
  weight: ["500"],
  subsets: ["latin"],
  variable: "--font-golos-text",
})

export const metadata: Metadata = {
  other: {
    "theme-color": "#FFFFFF",
    "msapplication-TileColor": "#FFFFFF",
    "apple-mobile-web-app-status-bar-style": "default",
  },
}

export default function AuthLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className={`${golosText.variable} flex min-h-screen flex-col overflow-x-hidden`}>
      <main className="flex-1">{children}</main>
      <AuthFooter />
      <AnalyticsConsentGate />
      <CookieBanner />
    </div>
  )
}
