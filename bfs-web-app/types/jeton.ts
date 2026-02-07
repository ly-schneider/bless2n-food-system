export type PosFulfillmentMode = "QR_CODE" | "JETON"

export interface Jeton {
  id: string
  name: string
  color: string
  usageCount?: number | null
}

export interface PosSettings {
  mode: PosFulfillmentMode
  missingJetons?: number
}
