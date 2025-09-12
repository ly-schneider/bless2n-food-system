"use client"

import { AlertCircle, Bell, CheckCircle, Clock, Eye, MoreHorizontal, RefreshCw, Timer } from "lucide-react"
import { useEffect, useState } from "react"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader } from "@/components/ui/card"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Separator } from "@/components/ui/separator"

// Orders now loaded from API

const statusConfig = {
  confirmed: {
    label: "New Order",
    color: "bg-blue-100 text-blue-800 border-blue-200",
    icon: Clock,
    action: "Start Preparing",
  },
  preparing: {
    label: "Preparing",
    color: "bg-yellow-100 text-yellow-800 border-yellow-200",
    icon: RefreshCw,
    action: "Mark Ready",
  },
  ready: {
    label: "Ready",
    color: "bg-green-100 text-green-800 border-green-200",
    icon: CheckCircle,
    action: "Complete Order",
  },
}

const typeConfig = {
  dine_in: { label: "Dine In", emoji: "🍽️" },
  takeout: { label: "Takeout", emoji: "🥡" },
  delivery: { label: "Delivery", emoji: "🚚" },
}

type OrderStatus = "confirmed" | "preparing" | "ready" | "completed"
type OrderType = "dine_in" | "takeout" | "delivery"
type Priority = "normal" | "high"
type OrderItem = { name: string; quantity: number }
type Order = {
  id: string
  orderNumber: string
  customer: { name: string; phone: string }
  status: OrderStatus
  type: OrderType
  tableNumber: string | null
  items: OrderItem[]
  total: number
  orderedAt: Date
  estimatedTime: number
  priority: Priority
}

export default function POSMainPage() {
  const [orders, setOrders] = useState<Order[]>([])
  const [currentTime, setCurrentTime] = useState(new Date())
  const [selectedOrder, setSelectedOrder] = useState<string | null>(null)

  // Update time every second for WebView
  useEffect(() => {
    const timer = setInterval(() => {
      setCurrentTime(new Date())
    }, 1000)

    return () => clearInterval(timer)
  }, [])

  // Load active orders from server
  useEffect(() => {
    const load = async () => {
      try {
        const res = await fetch("/api/pos/active-orders", { cache: "no-store" })
        if (!res.ok) throw new Error("Failed to fetch orders")
        const data = (await res.json()) as { orders?: ApiOrder[] }
        type ApiOrder = {
          id: string
          orderNumber?: string
          customer?: { name?: string; phone?: string }
          status?: string
          type?: string
          tableNumber?: string | null
          items?: { name?: string; quantity?: number }[]
          total?: number
          orderedAt?: string | Date
          estimatedTime?: number
          priority?: string
        }
        const mapped: Order[] = (data.orders as ApiOrder[] | undefined)?.map((o) => ({
          id: o.id,
          orderNumber: o.orderNumber || o.id,
          customer: { name: o.customer?.name || "Customer", phone: o.customer?.phone || "" },
          status: (o.status as OrderStatus) || "confirmed",
          type: (o.type as OrderType) || "takeout",
          tableNumber: o.tableNumber ?? null,
          items: (o.items || []).map((it) => ({ name: it.name || "", quantity: it.quantity || 0 })),
          total: Number(o.total || 0),
          orderedAt: o.orderedAt ? new Date(o.orderedAt) : new Date(),
          estimatedTime: Number(o.estimatedTime || 0),
          priority: (o.priority as Priority) || "normal",
        })) || []
        setOrders(mapped)
      } catch (e) {
        console.error(e)
      }
    }
    load()
  }, [])

  const handleStatusUpdate = (orderId: string, newStatus: OrderStatus) => {
    setOrders(
      (prevOrders) =>
        prevOrders
          .map((order) => (order.id === orderId ? { ...order, status: newStatus } : order))
          .filter((order) => order.status !== "completed") // Remove completed orders
    )
  }

  const getElapsedTime = (orderedAt: Date) => {
    const elapsed = Math.floor((currentTime.getTime() - orderedAt.getTime()) / 1000 / 60)
    return elapsed
  }

  return (
    <div className="h-full space-y-4 p-4">
      {/* Dashboard Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">Active Orders</h1>
          <p className="text-muted-foreground">
            {orders.length} active orders • {currentTime.toLocaleTimeString()}
          </p>
        </div>

        <div className="flex items-center space-x-2">
          <Button variant="outline" size="sm">
            <RefreshCw className="mr-2 h-4 w-4" />
            Refresh
          </Button>
          <Button variant="outline" size="sm">
            <Bell className="h-4 w-4" />
          </Button>
        </div>
      </div>

      {/* Order Stats */}
      <div className="grid grid-cols-3 gap-4">
        <Card>
          <CardContent className="p-4 text-center">
            <div className="text-2xl font-bold text-blue-600">
              {orders.filter((o) => o.status === "confirmed").length}
            </div>
            <div className="text-muted-foreground text-sm">New Orders</div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="p-4 text-center">
            <div className="text-2xl font-bold text-yellow-600">
              {orders.filter((o) => o.status === "preparing").length}
            </div>
            <div className="text-muted-foreground text-sm">Preparing</div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="p-4 text-center">
            <div className="text-2xl font-bold text-green-600">{orders.filter((o) => o.status === "ready").length}</div>
            <div className="text-muted-foreground text-sm">Ready</div>
          </CardContent>
        </Card>
      </div>

      {/* Orders Grid - optimized for touch interaction */}
      <ScrollArea className="h-[calc(100vh-280px)]">
        <div className="grid grid-cols-1 gap-4 lg:grid-cols-2 xl:grid-cols-3">
          {orders.map((order) => {
            const statusInfo = statusConfig[order.status as keyof typeof statusConfig]
            const typeInfo = typeConfig[order.type as keyof typeof typeConfig]
            const StatusIcon = statusInfo.icon
            const elapsed = getElapsedTime(order.orderedAt)
            const isOvertime = elapsed > 30 // Flag orders over 30 minutes

            return (
              <Card
                key={order.id}
                className={`relative cursor-pointer transition-all hover:shadow-lg ${
                  selectedOrder === order.id ? "ring-primary ring-2" : ""
                } ${isOvertime ? "border-red-300" : ""}`}
                onClick={() => setSelectedOrder(selectedOrder === order.id ? null : order.id)}
              >
                <CardHeader className="pb-3">
                  <div className="flex items-center justify-between">
                    <div className="flex items-center space-x-2">
                      <span className="text-lg font-bold">{order.orderNumber}</span>
                      {order.priority === "high" && (
                        <Badge variant="destructive" className="text-xs">
                          URGENT
                        </Badge>
                      )}
                    </div>

                    <div className="flex items-center space-x-2">
                      <Badge variant="outline" className={statusInfo.color}>
                        <StatusIcon className="mr-1 h-3 w-3" />
                        {statusInfo.label}
                      </Badge>
                    </div>
                  </div>

                  <div className="flex items-center justify-between text-sm">
                    <div className="flex items-center space-x-2">
                      <span>{typeInfo.emoji}</span>
                      <span className="font-medium">{order.customer.name}</span>
                      {order.tableNumber && (
                        <Badge variant="outline" className="text-xs">
                          {order.tableNumber}
                        </Badge>
                      )}
                    </div>

                    <div className={`flex items-center ${isOvertime ? "text-red-600" : "text-muted-foreground"}`}>
                      <Timer className="mr-1 h-3 w-3" />
                      <span>{elapsed}m</span>
                    </div>
                  </div>
                </CardHeader>

                <CardContent className="space-y-3 pt-0">
                  {/* Order Items */}
                  <div className="space-y-1">
                    {order.items.map((item, index) => (
                      <div key={index} className="flex justify-between text-sm">
                        <span>
                          {item.quantity}x {item.name}
                        </span>
                      </div>
                    ))}
                  </div>

                  <Separator />

                  <div className="flex items-center justify-between">
                    <div className="text-lg font-bold">${order.total.toFixed(2)}</div>

                    {order.estimatedTime > 0 && (
                      <div className="text-muted-foreground text-sm">~{order.estimatedTime}min left</div>
                    )}
                  </div>

                  {/* Action Buttons - touch-optimized */}
                  <div className="flex space-x-2 pt-2">
                    <Button
                      size="sm"
                      className="flex-1"
                      onClick={(e) => {
                        e.stopPropagation()
                        handleStatusUpdate(order.id, getNextStatus(order.status))
                      }}
                    >
                      {statusInfo.action}
                    </Button>

                    <Button variant="outline" size="sm">
                      <Eye className="h-4 w-4" />
                    </Button>

                    <Button variant="outline" size="sm">
                      <MoreHorizontal className="h-4 w-4" />
                    </Button>
                  </div>
                </CardContent>

                {/* Overtime indicator */}
                {isOvertime && (
                  <div className="absolute top-2 right-2">
                    <AlertCircle className="h-5 w-5 text-red-500" />
                  </div>
                )}
              </Card>
            )
          })}
        </div>

        {orders.length === 0 && (
          <div className="py-12 text-center">
            <CheckCircle className="mx-auto mb-4 h-16 w-16 text-green-500" />
            <h3 className="text-lg font-semibold">All caught up!</h3>
            <p className="text-muted-foreground">No active orders at the moment.</p>
          </div>
        )}
      </ScrollArea>
    </div>
  )
}

function getNextStatus(currentStatus: OrderStatus): OrderStatus {
  switch (currentStatus) {
    case "confirmed":
      return "preparing"
    case "preparing":
      return "ready"
    case "ready":
      return "completed"
    default:
      return currentStatus
  }
}
