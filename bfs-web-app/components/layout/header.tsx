"use client"

import { LayoutList, LogOut, TextAlignEnd, User } from "lucide-react"
import Image from "next/image"
import Link from "next/link"
import { usePathname, useRouter } from "next/navigation"
import { Button } from "@/components/ui/button"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem, DropdownMenuTrigger
} from "@/components/ui/dropdown-menu"
import { useAuth } from "@/contexts/auth-context"

export default function Header() {
  const { accessToken, signOut } = useAuth()
  const router = useRouter()
  const pathname = usePathname()
  return (
    <header className="my-2 w-full">
      <div className={`mx-auto px-4 ${pathname.includes("/food/orders") || pathname.includes("/food/checkout") || pathname.includes("/profile") ? "max-w-xl" : "container"}`}>
        <div className="relative flex items-center justify-between">
          <Link href="/food" className="flex items-center">
            <div className="flex h-12 w-12 items-center justify-center rounded-full">
              <Image
                src="/assets/images/blessthun.png"
                alt="BlessThun Logo"
                width={48}
                height={48}
                className="h-12 w-12"
              />
            </div>
          </Link>

          <h1 className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 text-3xl font-bold">FOOD</h1>

          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="secondary" size="icon" className="h-12 w-12 rounded-full">
                <TextAlignEnd className="size-5.5 h-5.5 w-5.5" />
                <span className="sr-only">Menü öffnen</span>
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end" className="min-w-48 rounded-2xl">
              <DropdownMenuItem asChild>
                <Link href="/food/orders" className="flex items-center gap-2">
                  <LayoutList className="h-4 w-4" />
                  <span>Bestellungen</span>
                </Link>
              </DropdownMenuItem>

              <DropdownMenuItem asChild>
                {accessToken ? (
                  <Link href="/profile" className="flex items-center gap-2">
                    <User className="h-4 w-4" />
                    <span>Benutzer</span>
                  </Link>
                ) : (
                  <Link href="/login" className="flex items-center gap-2">
                    <User className="h-4 w-4" />
                    <span>Benutzer</span>
                  </Link>
                )}
              </DropdownMenuItem>
              {accessToken && (
                <DropdownMenuItem
                  className="flex items-center gap-2"
                  onClick={async () => {
                    await signOut()
                    router.push("/")
                  }}
                >
                  <LogOut className="h-4 w-4" />
                  <span>Abmelden</span>
                </DropdownMenuItem>
              )}

              {/* Rechtliches links moved to footer */}
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      </div>
    </header>
  )
}
