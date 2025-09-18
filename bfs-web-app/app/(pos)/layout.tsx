import { Metadata } from "next"

export const metadata: Metadata = {
  title: "POS - Bless2n Food System",
  description: "Point of Sale system for Bless2n Food",
}

export default function POSLayout({ children }: { children: React.ReactNode }) {
  return <div className="min-h-screen">{children}</div>
}
