// Scans must accept both forms: nanoids on new rows and legacy UUIDs on rows
// created before the cutover. ENTITY_ID_RE mirrors the backend internal/id alphabet.
export const ENTITY_ID_RE = /^[1-9A-HKMNP-Za-hkmnp-z_-]{12}$/

const UUID_RE = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i

const CAMPAIGN_PREFIX = "CAMP:"
// Order QR codes encode a pickup URL `${origin}/o/<orderId>`; extract the id
// from the path segment rather than matching a window anywhere in the URL.
const ORDER_URL_RE =
  /\/o\/([1-9A-HKMNP-Za-hkmnp-z_-]{12}|[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12})\/?$/i

function isEntityId(s: string): boolean {
  return ENTITY_ID_RE.test(s) || UUID_RE.test(s)
}

export type ParsedScan = { kind: "order" | "campaign"; id: string }

export function parseScan(raw: string): ParsedScan | null {
  const trimmed = (raw ?? "").trim()
  if (!trimmed) return null

  if (trimmed.toUpperCase().startsWith(CAMPAIGN_PREFIX)) {
    const token = trimmed.slice(CAMPAIGN_PREFIX.length).trim()
    return isEntityId(token) ? { kind: "campaign", id: token } : null
  }

  const orderId = trimmed.match(ORDER_URL_RE)?.[1]
  if (orderId) return { kind: "order", id: orderId }

  return isEntityId(trimmed) ? { kind: "order", id: trimmed } : null
}
