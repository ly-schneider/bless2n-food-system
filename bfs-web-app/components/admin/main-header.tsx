"use client"
import { Home, LogOut, TextAlignEnd, User } from "lucide-react"
import Image from "next/image"
import Link from "next/link"
import { useRouter } from "next/navigation"
import { Button } from "@/components/ui/button"
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from "@/components/ui/dropdown-menu"
import { cn } from "@/lib/utils"

export function AdminMainHeader({ className }: { className?: string }) {
  const router = useRouter()

  return (
    <div className={cn("bg-background sticky top-0 z-[50]", className)}>
      <div className="mx-auto w-full px-6 md:px-8 lg:px-10">
        <div className="flex h-16 items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="bg-muted flex h-7 w-7 items-center justify-center overflow-hidden rounded-full">
              <Image src="/assets/images/blessthun.png" alt="BlessThun" width={28} height={28} />
            </div>
            <span className="text-base font-semibold">BlessThun</span>
          </div>
          <div className="flex items-center gap-2">
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="outline" size="sm" aria-haspopup="menu" className="rounded-[11px] border-0 py-4.5">
                  <TextAlignEnd className="size-4" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end" className="min-w-48 rounded-2xl">
                <DropdownMenuItem asChild>
                  <Link href="/profile" className="flex items-center gap-2">
                    <User className="h-4 w-4" />
                    <span>Benutzer</span>
                  </Link>
                </DropdownMenuItem>
                <DropdownMenuItem asChild>
                  <Link href="/food" className="flex items-center gap-2">
                    <Home className="h-4 w-4" />
                    <span>Food</span>
                  </Link>
                </DropdownMenuItem>
                <DropdownMenuItem
                  className="flex items-center gap-2"
                  onClick={async () => {
                    router.push("/")
                  }}
                >
                  <LogOut className="h-4 w-4" />
                  <span>Abmelden</span>
                </DropdownMenuItem>

                {/* Rechtliches links moved to footer */}
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
        </div>
      </div>
    </div>
  )
}
