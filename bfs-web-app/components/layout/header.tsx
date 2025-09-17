import { QrCode } from "lucide-react"
import Link from "next/link"
import { IconLink } from "@/components/ui/icon-button"

export default function Header() {
  return (
    <header className="w-full my-2">
      <div className="container mx-auto px-4">
        <div className="relative flex items-center justify-between">
          {/* Left: Logo */}
          <Link href="/" className="flex items-center">
            <div className="flex sm:h-16 sm:w-16 h-12 w-12 items-center justify-center rounded-full">
              <img src="/assets/images/blessthun.png" alt="BlessThun Logo" className="sm:h-16 sm:w-16 h-12 w-12" />
            </div>
          </Link>

          {/* Center: Title - Absolutely centered */}
          <h1 className="absolute left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2 text-3xl font-bold">FOOD</h1>

          {/* Right: Orders Button */}
          <IconLink href="/orders" variant="secondary" size="xl" shape="circle" className="sm:w-16 sm:h-16 w-12 h-12">
            <QrCode className="sm:h-8 sm:w-8 h-6 w-6" />
          </IconLink>
        </div>
      </div>
    </header>
  )
}