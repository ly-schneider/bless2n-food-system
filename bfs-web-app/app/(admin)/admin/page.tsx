import { AlertCircle, CheckCircle, Clock, DollarSign, ShoppingCart, TrendingUp, Users } from "lucide-react"
import { Metadata } from "next"
import { Suspense } from "react"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import OrderAPI from "@/lib/api/orders"

export const metadata: Metadata = {
  title: "Admin Dashboard - Bless2n Food System",
  description: "Admin dashboard for managing the food ordering system",
}

type RecentOrder = {
  id: string
  customer: string
  status: string
  total: number
  time: string
  items: number
}

type RawOrder = {
  id: string
  customer?: { name?: string }
  contactEmail?: string
  status?: string
  total?: number
  createdAt?: string
  items?: unknown[]
}

async function loadRecentOrders(): Promise<RecentOrder[]> {
  const data = await OrderAPI.listOrders({ limit: 10 })
  return (data.orders || []).map((o: RawOrder) => ({
    id: o.id,
    customer: o.customer?.name || o.contactEmail || "Customer",
    status: o.status || "pending",
    total: Number(o.total || 0),
    time: o.createdAt ? new Date(o.createdAt).toLocaleTimeString() : "",
    items: Array.isArray(o.items) ? o.items.length : 0,
  }))
}

const orderStatusConfig = {
  pending: { label: "Pending", variant: "secondary" as const, icon: Clock },
  preparing: { label: "Preparing", variant: "default" as const, icon: Clock },
  ready: { label: "Ready", variant: "default" as const, icon: CheckCircle },
  completed: { label: "Completed", variant: "default" as const, icon: CheckCircle },
  cancelled: { label: "Cancelled", variant: "destructive" as const, icon: AlertCircle },
}

export default async function AdminDashboard() {
  const recentOrders = await loadRecentOrders()
  const totalRevenue = recentOrders.reduce((sum, o) => sum + o.total, 0)
  const orderCount = recentOrders.length
  const avgOrderValue = orderCount ? totalRevenue / orderCount : 0

  return (
    <div className="space-y-6">
      {/* Page Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">Dashboard</h1>
          <p className="text-muted-foreground">Welcome back! Here's what's happening with your restaurant today.</p>
        </div>
        <div className="text-muted-foreground text-right text-sm">Last updated: {new Date().toLocaleTimeString()}</div>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-1 gap-6 md:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Total Revenue</CardTitle>
            <DollarSign className="text-muted-foreground h-4 w-4" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">${totalRevenue.toFixed(2)}</div>
            <div className="text-muted-foreground text-xs">from recent orders</div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Orders</CardTitle>
            <ShoppingCart className="text-muted-foreground h-4 w-4" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{orderCount}</div>
            <div className="text-muted-foreground text-xs">in recent activity</div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Avg Order Value</CardTitle>
            <TrendingUp className="text-muted-foreground h-4 w-4" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">${avgOrderValue.toFixed(2)}</div>
            <div className="text-muted-foreground text-xs">recent orders</div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Active Users</CardTitle>
            <Users className="text-muted-foreground h-4 w-4" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">--</div>
            <div className="text-muted-foreground text-xs">not available</div>
          </CardContent>
        </Card>
      </div>

      {/* Main Content Grid */}
      <div className="grid gap-6 lg:grid-cols-3">
        {/* Recent Orders */}
        <div className="lg:col-span-2">
          <Card>
            <CardHeader className="flex flex-row items-center justify-between">
              <CardTitle>Recent Orders</CardTitle>
              <Button variant="outline" size="sm">
                View All Orders
              </Button>
            </CardHeader>
            <CardContent>
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Order ID</TableHead>
                    <TableHead>Customer</TableHead>
                    <TableHead>Status</TableHead>
                    <TableHead>Items</TableHead>
                    <TableHead>Total</TableHead>
                    <TableHead>Time</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {recentOrders.map((order) => {
                    const statusConfig = orderStatusConfig[order.status as keyof typeof orderStatusConfig]
                    const StatusIcon = statusConfig.icon

                    return (
                      <TableRow key={order.id}>
                        <TableCell className="font-medium">{order.id}</TableCell>
                        <TableCell>{order.customer}</TableCell>
                        <TableCell>
                          <Badge variant={statusConfig.variant} className="flex w-fit items-center">
                            <StatusIcon className="mr-1 h-3 w-3" />
                            {statusConfig.label}
                          </Badge>
                        </TableCell>
                        <TableCell>{order.items} items</TableCell>
                        <TableCell>${order.total.toFixed(2)}</TableCell>
                        <TableCell className="text-muted-foreground">{order.time}</TableCell>
                      </TableRow>
                    )
                  })}
                </TableBody>
              </Table>
            </CardContent>
          </Card>
        </div>

        {/* Quick Actions */}
        <div className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle>Quick Actions</CardTitle>
            </CardHeader>
            <CardContent className="space-y-3">
              <Button className="w-full" variant="default">
                New Menu Item
              </Button>
              <Button className="w-full" variant="outline">
                Manage Categories
              </Button>
              <Button className="w-full" variant="outline">
                View Analytics
              </Button>
              <Button className="w-full" variant="outline">
                User Management
              </Button>
            </CardContent>
          </Card>

          {/* System Status */}
          <Card>
            <CardHeader>
              <CardTitle>System Status</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <Suspense fallback={<div>Loading system status...</div>}>
                <SystemStatus />
              </Suspense>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  )
}

async function SystemStatus() {
  // TODO: Replace with actual system health checks
  const systemStatus = {
    database: "healthy",
    paymentGateway: "healthy",
    emailService: "warning",
    storage: "healthy",
  }

  const getStatusColor = (status: string) => {
    switch (status) {
      case "healthy":
        return "text-green-600"
      case "warning":
        return "text-yellow-600"
      case "error":
        return "text-red-600"
      default:
        return "text-muted-foreground"
    }
  }

  const getStatusIcon = (status: string) => {
    switch (status) {
      case "healthy":
        return CheckCircle
      case "warning":
        return AlertCircle
      case "error":
        return AlertCircle
      default:
        return Clock
    }
  }

  return (
    <div className="space-y-3">
      {Object.entries(systemStatus).map(([service, status]) => {
        const StatusIcon = getStatusIcon(status)
        return (
          <div key={service} className="flex items-center justify-between">
            <span className="text-sm capitalize">{service.replace(/([A-Z])/g, " $1")}</span>
            <div className={`flex items-center ${getStatusColor(status)}`}>
              <StatusIcon className="mr-1 h-4 w-4" />
              <span className="text-sm capitalize">{status}</span>
            </div>
          </div>
        )
      })}
    </div>
  )
}
