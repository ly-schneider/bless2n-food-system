export type VolunteerCampaignStatus = "draft" | "active" | "ended"

export interface VolunteerCampaignSummary {
  id: string
  claimToken: string
  name: string
  accessCode: string
  validFrom?: string | null
  validUntil?: string | null
  status: VolunteerCampaignStatus
  maxRedemptions: number
  redemptionCount: number
  createdAt: string
  updatedAt: string
}

export interface VolunteerCampaignProductItem {
  productId: string
  productName: string
  productImage?: string | null
  quantity: number
}

export interface VolunteerCampaignRedemptionItem {
  id: string
  orderId: string
  createdAt: string
}

export interface VolunteerCampaignDetail extends VolunteerCampaignSummary {
  products: VolunteerCampaignProductItem[]
  redemptions: VolunteerCampaignRedemptionItem[]
}

export interface ClaimCampaignPublic {
  name: string
  validFrom?: string | null
  validUntil?: string | null
  status: VolunteerCampaignStatus
}

export interface ClaimCampaignResponse {
  campaign: ClaimCampaignPublic
  products: VolunteerCampaignProductItem[]
  qrPayload: string
}
