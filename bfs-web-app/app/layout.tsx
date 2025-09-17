import "styles/tailwind.css"
import type { Metadata } from "next"
import { Golos_Text } from "next/font/google"

const golosText = Golos_Text({
  weight: ["500"],
  subsets: ["latin"],
  variable: "--font-golos-text",
})

export const metadata: Metadata = {
  other: {
    "theme-color": "#E9E7E6",
    "msapplication-TileColor": "#E9E7E6",
    "apple-mobile-web-app-status-bar-style": "default",
  },
}

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body className={`${golosText.variable}`}>{children}</body>
    </html>
  )
}
