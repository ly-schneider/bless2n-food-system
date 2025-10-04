"use client"
import Link from "next/link"
import { usePathname } from "next/navigation"

export type AdminNavItem = { href: string; label: string }

export function AdminNav({ items }: { items: AdminNavItem[] }) {
  const pathname = usePathname()
  return (
    <nav className="flex gap-4 border-b border-gray-200 px-4 py-2 text-sm sticky top-0 bg-white z-10">
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

