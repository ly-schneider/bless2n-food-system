export type VolunteerCampaignStatus = "draft" | "active" | "ended"

export interface VolunteerCampaignSummary {
  id: string
  claimToken: string
  name: string
  accessCode: string
  validFrom?: string | null
  validUntil?: string | null
  status: VolunteerCampaignStatus
  totalSlots: number
  redeemedSlots: number
  reservedSlots: number
  createdAt: string
  updatedAt: string
}

export interface VolunteerCampaignProductItem {
  productId: string
  productName: string
  quantity: number
}

export interface VolunteerCampaignSlotItem {
  id: string
  orderId: string
  reservedBySession?: string | null
  reservedUntil?: string | null
  isRedeemed: boolean
  redeemedAt?: string | null
}

export interface VolunteerCampaignDetail extends VolunteerCampaignSummary {
  products: VolunteerCampaignProductItem[]
  slots: VolunteerCampaignSlotItem[]
}

export interface ClaimCampaignPublic {
  name: string
  validFrom?: string | null
  validUntil?: string | null
}

export interface ClaimSlotLineBrief {
  productName: string
  productImage?: string | null
  quantity: number
}

export interface ClaimSlotSummary {
  id: string
  orderId: string
  reservedUntil?: string | null
  lines: ClaimSlotLineBrief[]
}

export interface ClaimListResponse {
  campaign: ClaimCampaignPublic
  totalSlots: number
  availableCount: number
  available: ClaimSlotSummary[]
  reservedByMe: ClaimSlotSummary[]
}

export interface ClaimSlotDetail {
  id: string
  orderId: string
  reservedUntil?: string | null
  isRedeemed: boolean
  redeemedAt?: string | null
  lines: ClaimSlotLineBrief[]
}
