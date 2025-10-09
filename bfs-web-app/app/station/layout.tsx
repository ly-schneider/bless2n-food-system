import { Metadata } from "next"
import Image from "next/image"

export const metadata: Metadata = {
  title: "Station Portal - BlessThun Food",
  description: "Station access and management portal",
}

export default function StationLayout({ children }: { children: React.ReactNode }) {
  return (
    <main className="container mx-auto mt-6 px-4">
      <div className="mx-auto flex items-center justify-center">
        <div className="inline-flex items-center gap-3">
          <Image src="/assets/images/blessthun.png" alt="BlessThun Logo" width={40} height={40} className="h-10 w-10" />
          <span className="text-lg font-semibold tracking-tight">BlessThun Food</span>
        </div>
      </div>
      <div>{children}</div>
    </main>
  )
}
