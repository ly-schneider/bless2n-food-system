// Small crypto helpers for server routes

export function randomUrlSafe(length: number): string {
  const bytes = crypto.getRandomValues(new Uint8Array(length))
  const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_"
  let out = ""
  for (let i = 0; i < bytes.length; i++) out += chars[bytes[i]! % chars.length]
  return out
}

export async function sha256Base64Url(input: string): Promise<string> {
  const enc = new TextEncoder().encode(input)
  const digest = await crypto.subtle.digest("SHA-256", enc)
  let str = ""
  const arr = new Uint8Array(digest)
  for (let i = 0; i < arr.length; i++) str += String.fromCharCode(arr[i]!)
  return btoa(str).replaceAll("+", "-").replaceAll("/", "_").replace(/=+$/, "")
}
