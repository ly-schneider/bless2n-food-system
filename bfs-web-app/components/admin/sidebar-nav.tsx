"use client"
import { Circle, CreditCard, Grid2x2, Home, KeyRound, MailPlus, MonitorCheck, ReceiptText, Users } from "lucide-react"
import Link from "next/link"
import { usePathname } from "next/navigation"

type NavItem = {
  href: string
  label: string
  icon: React.ReactNode
  badge?: number
}

export function AdminSidebar({ badges }: { badges?: Partial<Record<string, number>> }) {
  const pathname = usePathname()
  const items: NavItem[] = [
    { href: "/admin", label: "Home", icon: <Home className="size-5" /> },
    { href: "/admin/menu", label: "Menu", icon: <Grid2x2 className="size-5" /> },
    { href: "/admin/orders", label: "Bestellungen", icon: <ReceiptText className="size-5" />, badge: badges?.orders },
    { href: "/admin/users", label: "Benutzer", icon: <Users className="size-5" /> },
    { href: "/admin/invites", label: "Einladungen", icon: <MailPlus className="size-5" />, badge: badges?.invites },
    { href: "/admin/sessions", label: "Sessions", icon: <KeyRound className="size-5" /> },
    { href: "/admin/stations", label: "Stationen", icon: <MonitorCheck className="size-5" /> },
    { href: "/admin/pos", label: "POS", icon: <CreditCard className="size-5" /> },
    { href: "/admin/jetons", label: "Jetons", icon: <Circle className="size-5" /> },
  ]

  return (
    <nav aria-label="Hauptnavigation" className="w-full">
      {/* Mobile: horizontal badges */}
      <div className="flex flex-wrap gap-2 px-6 md:hidden md:px-8 lg:px-10">
        {items.map((it) => {
          const active = pathname === it.href
          return (
            <Link
              key={it.href}
              href={it.href}
              className={`inline-flex h-9 items-center gap-2 rounded-full border px-3 ${
                active ? "bg-muted font-semibold" : "bg-card"
              }`}
              aria-current={active ? "page" : undefined}
            >
              {it.icon}
              <span className="text-sm">{it.label}</span>
              {typeof it.badge === "number" && it.badge > 0 && (
                <span className="bg-destructive rounded-full px-1.5 py-0.5 text-xs text-white">{it.badge}</span>
              )}
            </Link>
          )
        })}
      </div>
      {/* Desktop: vertical rounded panel with margin spacing from left */}
      <div className="hidden md:block">
        <div className="bg-card ml-6 max-w-md rounded-[11px] p-1 shadow-sm md:ml-8 lg:ml-10">
          {items.map((it) => {
            const active = pathname === it.href
            return (
              <Link
                key={it.href}
                href={it.href}
                className={`hover:bg-accent-hover flex h-10 items-center justify-between gap-2 rounded-lg px-3 ${
                  active ? "bg-muted font-medium" : ""
                }`}
                aria-current={active ? "page" : undefined}
              >
                <span className="flex items-center gap-2.5 text-sm">
                  {it.icon}
                  {it.label}
                </span>
                {typeof it.badge === "number" && it.badge > 0 && (
                  <span className="bg-destructive rounded-full px-1.5 py-0.5 text-xs text-white">{it.badge}</span>
                )}
              </Link>
            )
          })}
        </div>
      </div>
    </nav>
  )
}

export function AdminShell({
  children,
  badges,
}: {
  children: React.ReactNode
  badges?: Partial<Record<string, number>>
}) {
  return (
    <div className="min-h-dvh w-full overflow-x-clip">
      <div className="mx-auto w-full">
        <div className="pt-3 md:hidden">
          <AdminSidebar badges={badges} />
        </div>
        <div className="hidden md:grid md:grid-cols-[300px_1fr] md:gap-6">
          <div>
            <AdminSidebar badges={badges} />
          </div>
          <div className="min-w-0 pr-6 md:pr-8 lg:pr-10">{children}</div>
        </div>
        <div className="min-w-0 px-6 pt-4 pb-10 md:hidden md:px-8 lg:px-10">{children}</div>
      </div>
    </div>
  )
}
