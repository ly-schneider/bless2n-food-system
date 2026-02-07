/**
 * Device authentication utilities for POS and Station devices.
 *
 * Device tokens are long-lived session tokens issued when a device is paired
 * by an admin. Revocation is handled server-side via the device_binding table.
 */

const DEVICE_TOKEN_KEY = "bfs.deviceToken"
const DEVICE_TYPE_KEY = "bfs.deviceType"

export type DeviceType = "pos" | "station"

export function getDeviceToken(): string | null {
  if (typeof window === "undefined") return null
  return localStorage.getItem(DEVICE_TOKEN_KEY)
}

export function setDeviceToken(token: string): void {
  if (typeof window === "undefined") return
  localStorage.setItem(DEVICE_TOKEN_KEY, token)
}

export function clearDeviceToken(): void {
  if (typeof window === "undefined") return
  localStorage.removeItem(DEVICE_TOKEN_KEY)
  localStorage.removeItem(DEVICE_TYPE_KEY)
}

export function getDeviceType(): DeviceType | null {
  if (typeof window === "undefined") return null
  return localStorage.getItem(DEVICE_TYPE_KEY) as DeviceType | null
}

export function setDeviceType(type: DeviceType): void {
  if (typeof window === "undefined") return
  localStorage.setItem(DEVICE_TYPE_KEY, type)
}

export function hasDeviceToken(): boolean {
  return getDeviceToken() !== null
}

interface DeviceTokenPayload {
  type: string
  sub: string
  device_type: DeviceType
  iat: number
}

/**
 * Parse device token claims without validation.
 * Note: This does NOT validate the token signature - use for display purposes only.
 */
export function parseDeviceTokenClaims(
  token: string
): { sub: string; device_type: DeviceType; iat: number } | null {
  try {
    const parts = token.split(".")
    if (parts.length !== 3) return null

    const payloadPart = parts[1]
    if (!payloadPart) return null

    const payload = JSON.parse(atob(payloadPart)) as DeviceTokenPayload
    if (payload.type !== "device") return null

    return {
      sub: payload.sub,
      device_type: payload.device_type,
      iat: payload.iat,
    }
  } catch {
    return null
  }
}
