import { Metadata } from "next"
import { VersionLabel } from "@/components/layout/version-label"

export const metadata: Metadata = {
  title: "POS - BlessThun Food",
  description: "Point of Sale system for Bless2n Food",
}

export default function POSLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="min-h-screen">
      {children}
      <VersionLabel className="fixed bottom-4 left-2" version={process.env.APP_VERSION} />
    </div>
  )
}
