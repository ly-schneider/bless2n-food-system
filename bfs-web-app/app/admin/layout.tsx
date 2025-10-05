import { Metadata } from "next"
import { AdminGuard } from "@/components/admin/guard"
import { AdminMainHeader } from "@/components/admin/main-header"
import { AdminShell } from "@/components/admin/sidebar-nav"

export const metadata: Metadata = {
  title: "Admin Portal - BlessThun Food",
  description: "Admin portal",
}

export default function AdminLayout({ children }: { children: React.ReactNode }) {
  return (
    <AdminGuard>
      <AdminMainHeader />
      <AdminShell>
        <div className="mx-auto w-full pb-10">{children}</div>
      </AdminShell>
    </AdminGuard>
  )
}
