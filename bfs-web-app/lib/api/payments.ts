import { apiRequest } from "../api"

type ApiEnvelope<T> = { data: T; message?: string }

export interface CheckoutItem {
  productId: string
  quantity: number
  configuration?: Record<string, string>
}

export interface CreateCheckoutRequest {
  items: CheckoutItem[]
  customerEmail?: string
  clientReferenceId?: string
}

export interface CreateCheckoutResponse {
  url: string
  sessionId: string
}

export async function createCheckoutSession(body: CreateCheckoutRequest) {
  const res = await apiRequest<ApiEnvelope<CreateCheckoutResponse>>(
    "/v1/payments/checkout",
    {
      method: "POST",
      body: JSON.stringify(body),
    }
  )
  return res.data
}
