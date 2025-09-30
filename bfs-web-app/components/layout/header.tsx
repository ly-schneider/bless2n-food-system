"use client"

import Link from "next/link"
import { Button } from "@/components/ui/button"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { EllipsisVertical, LayoutList } from "lucide-react"

export default function Header() {
  return (
    <header className="w-full my-2">
      <div className="container mx-auto px-4">
        <div className="relative flex items-center justify-between">
          {/* Left: Logo */}
          <Link href="/" className="flex items-center">
            <div className="flex h-12 w-12 items-center justify-center rounded-full">
              <img src="/assets/images/blessthun.png" alt="BlessThun Logo" className="h-12 w-12" />
            </div>
          </Link>

          {/* Center: Title - Absolutely centered */}
          <h1 className="absolute left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2 text-3xl font-bold">FOOD</h1>

          {/* Right: General menu with dropdown */}
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="secondary" size="icon" className="h-12 w-12 rounded-full">
                <EllipsisVertical className="h-5.5 w-5.5 size-5.5" />
                <span className="sr-only">Menü öffnen</span>
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end" className="min-w-48 rounded-2xl">
              <DropdownMenuLabel>Allgemein</DropdownMenuLabel>
              <DropdownMenuItem asChild>
                <Link href="/orders" className="flex items-center gap-2">
                  <LayoutList className="h-4 w-4" />
                  <span>Bestellungen</span>
                </Link>
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      </div>
    </header>
  )
}
