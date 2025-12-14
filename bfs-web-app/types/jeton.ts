export type PosFulfillmentMode = "QR_CODE" | "JETON"

export interface Jeton {
  id: string
  name: string
  paletteColor: string
  hexColor?: string | null
  colorHex: string
  usageCount?: number | null
}

export interface PosSettings {
  mode: PosFulfillmentMode
  missingJetons?: number
}
