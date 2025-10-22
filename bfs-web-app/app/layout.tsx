import "styles/tailwind.css"
import { Golos_Text } from "next/font/google"
import CookieBanner from "@/components/cookie-banner"
import AnalyticsConsentGate from "@/components/google-analytics"
import { AuthProvider } from "@/contexts/auth-context"

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
        <AnalyticsConsentGate gaId={process.env.NEXT_PUBLIC_GA_MEASUREMENT_ID} />
        <CookieBanner />
      </body>
    </html>
  )
}
