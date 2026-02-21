import { apiRequest } from "@/lib/api"

type SystemStatus = { enabled: boolean }

export async function getSystemStatus(): Promise<SystemStatus> {
  try {
    return await apiRequest<SystemStatus>("/v1/system/status")
  } catch {
    return { enabled: true }
  }
}
