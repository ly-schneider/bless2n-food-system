export async function readErrorMessage(res: Response): Promise<string> {
  try {
    const j = (await res.json()) as unknown
    if (j && typeof j === "object" && j !== null) {
      const obj = j as { message?: unknown; detail?: unknown }
      if (typeof obj.message === "string") return obj.message
      if (typeof obj.detail === "string") return obj.detail
    }
  } catch {}
  return `HTTP ${res.status}: ${res.statusText}`
}
