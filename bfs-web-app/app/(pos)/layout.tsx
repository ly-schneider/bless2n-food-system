import { AlertTriangle, CreditCard, LogOut, type LucideIcon, Printer, QrCode, ShoppingCart } from "lucide-react"
import { redirect } from "next/navigation"
import { Suspense } from "react"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { AuthService } from "@/lib/auth"
import { RBACService } from "@/lib/rbac"

interface POSLayoutProps {
  children: React.ReactNode
}

export default async function POSLayout({ children }: POSLayoutProps) {
  // Server-side auth check
  const user = await AuthService.getCurrentUser()

  if (!user) {
    redirect("/login?error=pos_required")
  }

  if (!user || !RBACService.canAccessStation(user.role)) {
    redirect("/?error=unauthorized")
  }

  return (
    <div className="bg-background h-screen overflow-hidden">
      {/* WebView-optimized header - minimal and touch-friendly */}
      <header className="bg-background flex h-16 items-center border-b px-4">
        <div className="flex flex-1 items-center space-x-4">
          <h1 className="text-xl font-bold">POS Terminal</h1>
          <Badge variant="outline" className="text-xs">
            Online
          </Badge>
        </div>

        {/* Current Time */}
        <div className="text-muted-foreground mr-4 text-sm">
          <Suspense fallback="--:--">
            <CurrentTime />
          </Suspense>
        </div>

        {/* User Info */}
        <div className="flex items-center space-x-2">
          <div className="text-right text-sm">
            <p className="font-medium">{user.name}</p>
            <p className="text-muted-foreground text-xs">POS #{1}</p>
          </div>
          <Button variant="ghost" size="sm">
            <LogOut className="h-4 w-4" />
          </Button>
        </div>
      </header>

      <div className="flex h-[calc(100vh-64px)]">
        {/* Quick Nav Sidebar - optimized for touch */}
        <aside className="bg-muted/30 flex w-20 flex-col border-r">
          <nav className="space-y-2 p-2">
            <POSNavButton href="/pos" icon={ShoppingCart} label="Orders" />
            <POSNavButton href="/pos/scan" icon={QrCode} label="Scan" />
            <POSNavButton href="/pos/payment" icon={CreditCard} label="Payment" />
            <POSNavButton href="/pos/receipt" icon={Printer} label="Receipt" />
          </nav>

          {/* Status Indicators */}
          <div className="mt-auto space-y-2 p-2">
            <Suspense fallback={<div className="h-8" />}>
              <SystemStatus />
            </Suspense>
          </div>
        </aside>

        {/* Main Content - fullscreen for WebView */}
        <main className="flex-1 overflow-auto">
          {/* Fallback error boundary for WebView issues */}
          <WebViewErrorBoundary>{children}</WebViewErrorBoundary>
        </main>
      </div>
    </div>
  )
}

function POSNavButton({ href, icon: Icon, label }: { href: string; icon: LucideIcon; label: string }) {
  return (
    <Button variant="ghost" size="sm" className="hover:bg-accent flex h-12 w-full flex-col p-2 text-xs" asChild>
      <a href={href}>
        {" "}
        {/* Use <a> instead of Next.js Link for WebView compatibility */}
        <Icon className="mb-1 h-5 w-5" />
        <span>{label}</span>
      </a>
    </Button>
  )
}

function CurrentTime() {
  // In real implementation, use a client component with useEffect for real-time updates
  const now = new Date()
  return <span>{now.toLocaleTimeString()}</span>
}

async function SystemStatus() {
  // Basic network reachability check to API
  let networkOk = false
  try {
    const base = process.env.NEXT_PUBLIC_API_BASE_URL || "http://localhost:8080"
    const res = await fetch(base, { method: "HEAD", cache: "no-store" })
    networkOk = !!res
  } catch {
    networkOk = false
  }
  const status = {
    printer: true, // integrate with device APIs if available
    payment: true, // integrate with gateway status endpoint when available
    network: networkOk,
  }

  return (
    <div className="space-y-1">
      <div
        className={`h-3 w-3 rounded-full ${status.printer ? "bg-green-500" : "bg-red-500"}`}
        title="Printer Status"
      />
      <div
        className={`h-3 w-3 rounded-full ${status.payment ? "bg-green-500" : "bg-red-500"}`}
        title="Payment Gateway"
      />
      <div
        className={`h-3 w-3 rounded-full ${status.network ? "bg-green-500" : "bg-red-500"}`}
        title="Network Status"
      />
    </div>
  )
}

// Error boundary specifically for WebView constraints
function WebViewErrorBoundary({ children }: { children: React.ReactNode }) {
  return (
    <div className="h-full">
      {children}
      {/* Fallback UI for WebView-specific issues */}
      <noscript>
        <div className="flex h-full items-center justify-center">
          <div className="space-y-4 text-center">
            <AlertTriangle className="mx-auto h-12 w-12 text-yellow-500" />
            <h2 className="text-lg font-semibold">JavaScript Required</h2>
            <p className="text-muted-foreground">Please enable JavaScript in your WebView settings.</p>
          </div>
        </div>
      </noscript>
    </div>
  )
}
