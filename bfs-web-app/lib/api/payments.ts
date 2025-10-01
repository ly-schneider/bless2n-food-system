import { apiRequest } from "../api"

type ApiEnvelope<T> = { data: T; message?: string }

export interface CheckoutItem {
  productId: string
  quantity: number
  configuration?: Record<string, string>
}

export interface CreateIntentRequest {
  items: CheckoutItem[]
  customerEmail?: string
  attemptId?: string
}

export interface CreateIntentResponse {
  clientSecret: string
  paymentIntentId: string
  orderId: string
}

export async function createPaymentIntent(body: CreateIntentRequest, accessToken?: string) {
  const res = await apiRequest<ApiEnvelope<CreateIntentResponse>>(
    "/v1/payments/create-intent",
    {
      method: "POST",
      body: JSON.stringify(body),
      headers: accessToken ? { Authorization: `Bearer ${accessToken}` } : undefined,
    }
  )
  return res.data
}

export async function attachReceiptEmail(paymentIntentId: string, email: string | null, accessToken?: string) {
  const res = await apiRequest<ApiEnvelope<{ paymentIntentId: string; receiptEmail?: string | null }>>(
    "/v1/payments/attach-email",
    {
      method: "PATCH",
      body: JSON.stringify({ paymentIntentId, email }),
      headers: accessToken ? { Authorization: `Bearer ${accessToken}` } : undefined,
    }
  )
  return res.data
}

export interface PaymentStatusResponse {
  id: string
  status: string
  amount: number
  currency: string
  chargeId?: string
  customer?: string
  receiptEmail?: string
  metadata?: Record<string, string>
}

export async function getPaymentStatus(id: string) {
  const res = await apiRequest<ApiEnvelope<PaymentStatusResponse>>(`/v1/payments/${id}`, {
    method: "GET",
  })
  return res.data
}
