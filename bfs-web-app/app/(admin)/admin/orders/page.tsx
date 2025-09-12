import { CheckCircle, Clock, Download, Eye, Filter, RefreshCw, Search, XCircle } from "lucide-react"
import { Metadata } from "next"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import OrderAPI from "@/lib/api/orders"

export const metadata: Metadata = {
  title: "Order Management - Admin Dashboard",
  description: "Manage and track all customer orders",
}

// UI-facing order types to avoid "any"
type OrderStatusKey = keyof typeof statusConfig
type OrderTypeKey = keyof typeof typeConfig

type UICustomer = {
  name?: string
  email?: string
  phone?: string
  isGuest?: boolean
}

type UIOrderItem = {
  name?: string
  quantity?: number
  price?: number
}

type UIOrder = {
  id: string
  orderNumber?: string
  customer?: UICustomer
  status?: string
  type?: string
  items: UIOrderItem[]
  total?: number
  tableNumber?: string | null
  orderedAt?: string
  estimatedTime?: number
}

const statusConfig = {
  pending: {
    label: "Pending",
    variant: "secondary" as const,
    icon: Clock,
    color: "bg-gray-100 text-gray-800",
  },
  confirmed: {
    label: "Confirmed",
    variant: "default" as const,
    icon: CheckCircle,
    color: "bg-blue-100 text-blue-800",
  },
  preparing: {
    label: "Preparing",
    variant: "default" as const,
    icon: RefreshCw,
    color: "bg-yellow-100 text-yellow-800",
  },
  ready: {
    label: "Ready",
    variant: "default" as const,
    icon: CheckCircle,
    color: "bg-green-100 text-green-800",
  },
  completed: {
    label: "Completed",
    variant: "outline" as const,
    icon: CheckCircle,
    color: "bg-green-100 text-green-800",
  },
  cancelled: {
    label: "Cancelled",
    variant: "destructive" as const,
    icon: XCircle,
    color: "bg-red-100 text-red-800",
  },
}

const typeConfig = {
  dine_in: { label: "Dine In", color: "bg-blue-100 text-blue-800" },
  takeout: { label: "Takeout", color: "bg-purple-100 text-purple-800" },
  delivery: { label: "Delivery", color: "bg-orange-100 text-orange-800" },
}

export default async function OrdersPage() {
  const data = await OrderAPI.listOrders({ limit: 50 })
  const orders: UIOrder[] = ((data.orders as unknown) as UIOrder[]) || []

  const orderCounts = {
    all: orders.length,
    pending: orders.filter((o) => o.status === "pending").length,
    preparing: orders.filter((o) => o.status === "preparing").length,
    ready: orders.filter((o) => o.status === "ready").length,
    completed: orders.filter((o) => o.status === "completed").length,
  }

  return (
    <div className="space-y-6">
      {/* Page Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">Order Management</h1>
          <p className="text-muted-foreground">Track and manage all customer orders in real-time</p>
        </div>
        <div className="flex items-center space-x-2">
          <Button variant="outline" size="sm">
            <Download className="mr-2 h-4 w-4" />
            Export
          </Button>
          <Button variant="outline" size="sm">
            <RefreshCw className="mr-2 h-4 w-4" />
            Refresh
          </Button>
        </div>
      </div>

      {/* Filters */}
      <Card>
        <CardContent className="p-4">
          <div className="flex flex-wrap gap-4">
            <div className="min-w-64 flex-1">
              <div className="relative">
                <Search className="text-muted-foreground absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2 transform" />
                <Input placeholder="Search by order ID, customer name, or phone..." className="pl-10" />
              </div>
            </div>

            <Select defaultValue="all">
              <SelectTrigger className="w-40">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">All Status</SelectItem>
                <SelectItem value="pending">Pending</SelectItem>
                <SelectItem value="preparing">Preparing</SelectItem>
                <SelectItem value="ready">Ready</SelectItem>
                <SelectItem value="completed">Completed</SelectItem>
              </SelectContent>
            </Select>

            <Select defaultValue="all">
              <SelectTrigger className="w-36">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">All Types</SelectItem>
                <SelectItem value="dine_in">Dine In</SelectItem>
                <SelectItem value="takeout">Takeout</SelectItem>
                <SelectItem value="delivery">Delivery</SelectItem>
              </SelectContent>
            </Select>

            <Button variant="outline" size="sm">
              <Filter className="mr-2 h-4 w-4" />
              More Filters
            </Button>
          </div>
        </CardContent>
      </Card>

      {/* Order Tabs */}
      <Tabs defaultValue="all" className="space-y-4">
        <TabsList className="grid w-full grid-cols-5">
          <TabsTrigger value="all" className="relative">
            All Orders
            <Badge variant="secondary" className="ml-2 text-xs">
              {orderCounts.all}
            </Badge>
          </TabsTrigger>
          <TabsTrigger value="pending">
            Pending
            <Badge variant="secondary" className="ml-2 text-xs">
              {orderCounts.pending}
            </Badge>
          </TabsTrigger>
          <TabsTrigger value="preparing">
            Preparing
            <Badge variant="secondary" className="ml-2 text-xs">
              {orderCounts.preparing}
            </Badge>
          </TabsTrigger>
          <TabsTrigger value="ready">
            Ready
            <Badge variant="secondary" className="ml-2 text-xs">
              {orderCounts.ready}
            </Badge>
          </TabsTrigger>
          <TabsTrigger value="completed">
            Completed
            <Badge variant="secondary" className="ml-2 text-xs">
              {orderCounts.completed}
            </Badge>
          </TabsTrigger>
        </TabsList>

        <TabsContent value="all">
          <OrdersTable orders={orders} />
        </TabsContent>

        <TabsContent value="pending">
          <OrdersTable orders={orders.filter((o) => o.status === "pending")} />
        </TabsContent>

        <TabsContent value="preparing">
          <OrdersTable orders={orders.filter((o) => o.status === "preparing")} />
        </TabsContent>

        <TabsContent value="ready">
          <OrdersTable orders={orders.filter((o) => o.status === "ready")} />
        </TabsContent>

        <TabsContent value="completed">
          <OrdersTable orders={orders.filter((o) => o.status === "completed")} />
        </TabsContent>
      </Tabs>
    </div>
  )
}

function OrdersTable({ orders }: { orders: UIOrder[] }) {
  return (
    <Card>
      <CardContent className="p-0">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Order</TableHead>
              <TableHead>Customer</TableHead>
              <TableHead>Type</TableHead>
              <TableHead>Status</TableHead>
              <TableHead>Items</TableHead>
              <TableHead>Total</TableHead>
              <TableHead>Time</TableHead>
              <TableHead className="text-right">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {orders.map((order) => {
              const statusInfo = statusConfig[(order.status as OrderStatusKey) || "confirmed"] || statusConfig.confirmed
              const typeInfo = typeConfig[(order.type as OrderTypeKey) || "takeout"]
              const StatusIcon = statusInfo.icon

              return (
                <TableRow key={order.id}>
                  <TableCell>
                    <div>
                      <div className="font-medium">{order.orderNumber || order.id}</div>
                      <div className="text-muted-foreground text-sm">{order.id}</div>
                    </div>
                  </TableCell>

                  <TableCell>
                    <div>
                      <div className="font-medium">{order.customer?.name || "Customer"}</div>
                      {order.customer?.email && (
                        <div className="text-muted-foreground text-sm">{order.customer.email}</div>
                      )}
                      {order.tableNumber && <div className="text-muted-foreground text-sm">{order.tableNumber}</div>}
                    </div>
                  </TableCell>

                  <TableCell>
                    <Badge variant="outline" className={typeInfo.color}>
                      {typeInfo.label}
                    </Badge>
                  </TableCell>

                  <TableCell>
                    <Badge variant={statusInfo.variant} className="flex w-fit items-center">
                      <StatusIcon className="mr-1 h-3 w-3" />
                      {statusInfo.label}
                    </Badge>
                  </TableCell>

                  <TableCell>
                    <div>
                      <div className="font-medium">{order.items.length} items</div>
                      <div className="text-muted-foreground text-sm">
                        {order.items[0]?.name}
                        {order.items.length > 1 && ` +${order.items.length - 1} more`}
                      </div>
                    </div>
                  </TableCell>

                  <TableCell className="font-medium">${Number(order.total || 0).toFixed(2)}</TableCell>

                  <TableCell>
                    {order.orderedAt && (
                      <div className="text-sm">{new Date(order.orderedAt).toLocaleTimeString()}</div>
                    )}
                    {(order.estimatedTime ?? 0) > 0 && (
                      <div className="text-muted-foreground text-sm">{order.estimatedTime} min remaining</div>
                    )}
                  </TableCell>

                  <TableCell className="text-right">
                    <div className="flex items-center justify-end space-x-2">
                      <Button variant="ghost" size="sm">
                        <Eye className="h-4 w-4" />
                      </Button>

                      {order.status !== "completed" && order.status !== "cancelled" && (
                        <Select defaultValue={order.status}>
                          <SelectTrigger className="w-32">
                            <SelectValue />
                          </SelectTrigger>
                          <SelectContent>
                            <SelectItem value="pending">Pending</SelectItem>
                            <SelectItem value="confirmed">Confirmed</SelectItem>
                            <SelectItem value="preparing">Preparing</SelectItem>
                            <SelectItem value="ready">Ready</SelectItem>
                            <SelectItem value="completed">Completed</SelectItem>
                            <SelectItem value="cancelled">Cancelled</SelectItem>
                          </SelectContent>
                        </Select>
                      )}
                    </div>
                  </TableCell>
                </TableRow>
              )
            })}
          </TableBody>
        </Table>

        {orders.length === 0 && (
          <div className="py-12 text-center">
            <div className="text-muted-foreground">No orders found</div>
          </div>
        )}
      </CardContent>
    </Card>
  )
}
