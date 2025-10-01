import { Metadata } from "next"

export const metadata: Metadata = {
  title: "Admin Portal - BlessThun Food",
  description: "Admin portal",
}

export default function AdminLayout({ children }: { children: React.ReactNode }) {
  return <div className="min-h-screen">{children}</div>
}
