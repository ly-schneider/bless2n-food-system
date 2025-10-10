"use client"
import Link from "next/link"
import { usePathname } from "next/navigation"

export type AdminNavItem = { href: string; label: string }

export function AdminNav({ items }: { items: AdminNavItem[] }) {
  const pathname = usePathname()
  return (
    <nav className="sticky top-0 z-10 flex gap-4 border-b border-gray-200 bg-white px-4 py-2 text-sm">
      {items.map((n) => {
        const active = pathname === n.href
        return (
          <Link key={n.href} href={n.href} className={active ? "font-semibold" : "text-gray-600 hover:text-black"}>
            {n.label}
          </Link>
        )
      })}
    </nav>
  )
}
