import { apiRequest } from "../api"

export interface CheckoutItem {
  productId: string
  quantity: number
  menuSelections?: { slotId: string; productId: string }[]
}

export interface CreateOrderRequest {
  items: CheckoutItem[]
  contactEmail?: string
}

export interface CreateOrderResponse {
  id: string
  status: string
  totalCents: number
}

export interface InitiatePaymentRequest {
  method: "twint"
  channel: "web"
  returnUrl: string
}

export interface InitiatePaymentResponse {
  orderId: string
  method: string
  redirectUrl?: string
  gatewayId?: number
}

export interface PaymentStatusResponse {
  id: string
  orderId: string
  status: "pending" | "paid" | "cancelled" | "refunded"
  method?: string
  amountCents?: number
}

export async function createOrder(body: CreateOrderRequest, accessToken?: string) {
  return apiRequest<CreateOrderResponse>("/v1/orders", {
    method: "POST",
    body: JSON.stringify(body),
    headers: accessToken ? { Authorization: `Bearer ${accessToken}` } : undefined,
  })
}

export async function initiatePayment(orderId: string, body: InitiatePaymentRequest, accessToken?: string) {
  return apiRequest<InitiatePaymentResponse>(`/v1/orders/${orderId}/payment`, {
    method: "POST",
    body: JSON.stringify(body),
    headers: accessToken ? { Authorization: `Bearer ${accessToken}` } : undefined,
  })
}

export async function getPaymentStatus(orderId: string) {
  return apiRequest<PaymentStatusResponse>(`/v1/orders/${orderId}/payment`, {
    method: "GET",
  })
}
