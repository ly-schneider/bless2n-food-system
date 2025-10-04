import { Metadata } from "next"

export const metadata: Metadata = {
  title: "Station Portal - BlessThun Food",
  description: "Station access and management portal",
}

export default function StationLayout({ children }: { children: React.ReactNode }) {
  return <div className="min-h-screen">{children}</div>
}
