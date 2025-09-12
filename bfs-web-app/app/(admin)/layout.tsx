import { BarChart3, Bell, FileText, LogOut, type LucideIcon, Menu, Settings, ShoppingCart, User, Users } from "lucide-react"
import Link from "next/link"
import { redirect } from "next/navigation"
import { Suspense } from "react"
import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"
import { Separator } from "@/components/ui/separator"
import { AuthService } from "@/lib/auth"
import { Permission, RBACService } from "@/lib/rbac"

interface AdminLayoutProps {
  children: React.ReactNode
}

export default async function AdminLayout({ children }: AdminLayoutProps) {
  // Server-side auth check
  const user = await AuthService.getCurrentUser()

  if (!user) {
    redirect("/login?error=admin_required")
  }

  if (!user || !RBACService.hasPermission(user.role, Permission.ADMIN_ANALYTICS)) {
    redirect("/?error=unauthorized")
  }

  return (
    <div className="bg-background min-h-screen">
      {/* Top Header */}
      <header className="bg-background/95 sticky top-0 z-50 border-b backdrop-blur">
        <div className="flex h-16 items-center px-6">
          <div className="flex items-center space-x-4">
            <Link href="/admin" className="text-xl font-bold">
              Admin Dashboard
            </Link>
          </div>

          <div className="ml-auto flex items-center space-x-4">
            <Button variant="ghost" size="sm">
              <Bell className="h-4 w-4" />
            </Button>

            <Separator orientation="vertical" className="h-6" />

            <div className="flex items-center space-x-2">
              <div className="text-right">
                <p className="text-sm font-medium">{user.name}</p>
                <p className="text-muted-foreground text-xs">{user.role}</p>
              </div>
              <Button variant="ghost" size="sm">
                <User className="h-4 w-4" />
              </Button>
            </div>

            <Button variant="ghost" size="sm">
              <LogOut className="h-4 w-4" />
            </Button>
          </div>
        </div>
      </header>

      <div className="flex">
        {/* Sidebar */}
        <aside className="bg-muted/40 min-h-screen w-64 border-r">
          <nav className="space-y-2 p-4">
            <div className="mb-4">
              <h3 className="text-muted-foreground mb-2 text-sm font-semibold tracking-wider uppercase">Dashboard</h3>
              <NavItem href="/admin" icon={BarChart3}>
                Overview
              </NavItem>
            </div>

            <div className="mb-4">
              <h3 className="text-muted-foreground mb-2 text-sm font-semibold tracking-wider uppercase">Management</h3>
              <NavItem href="/admin/orders" icon={ShoppingCart}>
                Orders
              </NavItem>
              <NavItem href="/admin/menu" icon={Menu}>
                Menu
              </NavItem>
              <NavItem href="/admin/users" icon={Users}>
                Users
              </NavItem>
            </div>

            <div className="mb-4">
              <h3 className="text-muted-foreground mb-2 text-sm font-semibold tracking-wider uppercase">Analytics</h3>
              <NavItem href="/admin/analytics" icon={BarChart3}>
                Reports
              </NavItem>
              <NavItem href="/admin/audit" icon={FileText}>
                Audit Logs
              </NavItem>
            </div>

            <div>
              <h3 className="text-muted-foreground mb-2 text-sm font-semibold tracking-wider uppercase">System</h3>
              <NavItem href="/admin/settings" icon={Settings}>
                Settings
              </NavItem>
            </div>
          </nav>

          {/* Quick Stats Card */}
          <div className="p-4">
            <Card>
              <CardContent className="p-4">
                <h4 className="mb-3 font-semibold">Quick Stats</h4>
                <Suspense fallback={<div className="space-y-2">Loading...</div>}>
                  <QuickStats />
                </Suspense>
              </CardContent>
            </Card>
          </div>
        </aside>

        {/* Main Content */}
        <main className="flex-1">
          <div className="p-6">{children}</div>
        </main>
      </div>
    </div>
  )
}

function NavItem({ href, icon: Icon, children }: { href: string; icon: LucideIcon; children: React.ReactNode }) {
  return (
    <Button variant="ghost" className="w-full justify-start" asChild>
      <Link href={href}>
        <Icon className="mr-3 h-4 w-4" />
        {children}
      </Link>
    </Button>
  )
}

async function QuickStats() {
  // TODO: Fetch real stats from API
  const stats = {
    pendingOrders: 8,
    todayRevenue: 1250.5,
    activeUsers: 145,
  }

  return (
    <div className="space-y-3">
      <div className="flex justify-between text-sm">
        <span className="text-muted-foreground">Pending Orders</span>
        <span className="font-semibold">{stats.pendingOrders}</span>
      </div>
      <div className="flex justify-between text-sm">
        <span className="text-muted-foreground">Today's Revenue</span>
        <span className="font-semibold">${stats.todayRevenue.toFixed(2)}</span>
      </div>
      <div className="flex justify-between text-sm">
        <span className="text-muted-foreground">Active Users</span>
        <span className="font-semibold">{stats.activeUsers}</span>
      </div>
    </div>
  )
}
