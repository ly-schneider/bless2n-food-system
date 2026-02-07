export function generateLocalId(): string {
  const timestamp = Date.now().toString(16).padStart(12, "0")
  const random = crypto.getRandomValues(new Uint8Array(8))
  const randomHex = Array.from(random)
    .map((b) => b.toString(16).padStart(2, "0"))
    .join("")
  return `${timestamp}${randomHex}`
}
