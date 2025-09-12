import { NextResponse } from "next/server"
import OrderAPI from "@/lib/api/orders"

// Normalize backend orders into POS-friendly shape
type RawOrder = {
  id: string
  orderNumber?: string
  customer?: { name?: string; phone?: string }
  status?: string
  type?: string
  tableNumber?: string | null
  items?: { name?: string; quantity?: number }[]
  total?: number
  orderedAt?: string
  estimatedTime?: number
  priority?: string
}

function toPOSOrder(o: RawOrder) {
  return {
    id: o.id,
    orderNumber: o.orderNumber || o.id,
    customer: { name: o.customer?.name || "Customer", phone: o.customer?.phone || "" },
    status: (o.status as string) || "confirmed",
    type: (o.type as string) || "takeout",
    tableNumber: o.tableNumber ?? null,
    items: Array.isArray(o.items) ? o.items.map((it) => ({ name: it.name, quantity: it.quantity })) : [],
    total: Number(o.total || 0),
    orderedAt: o.orderedAt ? new Date(o.orderedAt) : new Date(),
    estimatedTime: Number(o.estimatedTime || 0),
    priority: (o.priority as string) || "normal",
  }
}

export async function GET() {
  try {
    const data = await OrderAPI.listOrders({ limit: 50 })
    const orders = (data.orders || [])
      .filter((o: RawOrder) => o.status !== "completed")
      .map(toPOSOrder)
    return NextResponse.json({ orders })
  } catch (err: unknown) {
    return NextResponse.json({ error: (err as Error)?.message || "Failed to load orders" }, { status: 500 })
  }
}
