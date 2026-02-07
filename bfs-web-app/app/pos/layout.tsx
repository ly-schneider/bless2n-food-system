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
      <VersionLabel className="fixed left-2 bottom-2" />
    </div>
  )
}
