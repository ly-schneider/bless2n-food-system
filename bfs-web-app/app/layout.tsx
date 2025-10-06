import "styles/tailwind.css"
import { AuthProvider } from "@/contexts/auth-context"
import { Golos_Text } from "next/font/google"
import AnalyticsConsentGate from "@/components/google-analytics"
import CookieBanner from "@/components/cookie-banner"

const golosText = Golos_Text({
  weight: ["500"],
  subsets: ["latin"],
  variable: "--font-golos-text",
})

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="de" className={golosText.variable}>
      <body>
        <AuthProvider>{children}</AuthProvider>
        <AnalyticsConsentGate />
        <CookieBanner />
      </body>
    </html>
  )
}
