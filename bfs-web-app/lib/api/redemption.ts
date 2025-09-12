import { GetOrderResponse, RedeemOrderItemsRequest, RedeemOrderItemsResponse } from "../../types/order"

const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL || "http://localhost:8080"

async function apiRequest<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
  const response = await fetch(`${API_BASE_URL}${endpoint}`, {
    ...options,
    headers: {
      "Content-Type": "application/json",
      ...options.headers,
    },
  })

  if (!response.ok) {
    const error = await response.text()
    throw new Error(error || `HTTP ${response.status}`)
  }

  const data = (await response.json()) as unknown
  return data as T
}

// Public Redemption API methods (no auth required for station devices)
export async function getOrderForRedemption(orderId: string): Promise<GetOrderResponse> {
  return apiRequest<GetOrderResponse>(`/v1/redemption/orders/${orderId}`)
}

export async function redeemOrderItems(request: RedeemOrderItemsRequest): Promise<RedeemOrderItemsResponse> {
  return apiRequest<RedeemOrderItemsResponse>("/v1/redemption/redeem", {
    method: "POST",
    body: JSON.stringify(request),
  })
}

const RedemptionAPI = {
  getOrderForRedemption,
  redeemOrderItems,
}

export default RedemptionAPI
