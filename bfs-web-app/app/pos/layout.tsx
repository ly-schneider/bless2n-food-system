import { Metadata } from "next"
import { VersionLabel } from "@/components/layout/version-label"

export const metadata: Metadata = {
  title: "POS",
  description: "Point-of-Sale Terminal für das BlessThun Food Bestellsystem.",
}

export default function POSLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="min-h-screen">
      {children}
      <VersionLabel className="fixed bottom-4 left-2" version={process.env.APP_VERSION} />
    </div>
  )
}
