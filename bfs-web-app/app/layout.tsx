import "styles/tailwind.css"
import { Golos_Text } from "next/font/google"

const golosText = Golos_Text({
  weight: ["500"],
  subsets: ["latin"],
  variable: "--font-golos-text",
})

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body className={`${golosText.variable}`}>{children}</body>
    </html>
  )
}
