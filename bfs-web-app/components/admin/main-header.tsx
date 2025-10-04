"use client"
import { useRouter } from "next/navigation"
import { Bell, ChevronDown, RefreshCw } from "lucide-react"
import Image from "next/image"
import { Button } from "@/components/ui/button"
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuLabel, DropdownMenuTrigger } from "@/components/ui/dropdown-menu"
import { cn } from "@/lib/utils"

export function AdminMainHeader({ className }: { className?: string }) {
  const router = useRouter()
  return (
    <div className={cn("sticky top-0 z-[50] bg-background", className)}>
      <div className="mx-auto w-full px-6 md:px-8 lg:px-10">
        <div className="h-16 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="h-7 w-7 rounded-full overflow-hidden bg-muted flex items-center justify-center">
              <Image src="/assets/images/blessthun.png" alt="BlessThun" width={28} height={28} />
            </div>
            <span className="text-base font-semibold">BlessThun</span>
          </div>
          <div className="flex items-center gap-2">
            <Button
              variant="primary"
              size="sm"
              onClick={() => {
                try {
                  window.dispatchEvent(new CustomEvent("admin:refresh"))
                } catch {}
              }}
              aria-label="Aktualisieren"
            >
              <RefreshCw className="size-4" />
              <span>Aktualisieren</span>
            </Button>
            <Button variant="ghost" size="icon" aria-label="Benachrichtigungen">
              <Bell className="size-5" />
            </Button>
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="outline" size="sm" aria-haspopup="menu">
                  Account <ChevronDown className="size-4" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end" className="rounded-xl">
                <DropdownMenuLabel>Konto</DropdownMenuLabel>
                <DropdownMenuItem onClick={() => router.push("/profile")}>Profil</DropdownMenuItem>
                <DropdownMenuItem onClick={() => router.push("/")}>Zur Startseite</DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
        </div>
      </div>
    </div>
  )
}
