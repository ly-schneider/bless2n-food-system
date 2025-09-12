import { Metadata } from "next"

export const metadata: Metadata = {
  title: "Partner Station Portal - Bless2n Food",
  description: "Partner station access and management portal",
}

export default function StationLayout({ children }: { children: React.ReactNode }) {
  return <div className="min-h-screen">{children}</div>
}
