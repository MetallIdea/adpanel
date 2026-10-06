import { get } from "../utils/requests"

export interface ServerStatus {
  cpu: {
    usage_percent: number
  }
  memory: {
    used_mb: number
    total_mb: number
    usage_percent: number
  }
  disk: {
    used_mb: number
    total_mb: number
    usage_percent: number
  }
}

export async function getServerStatus(): Promise<ServerStatus> {
  return get<ServerStatus>("/server/status")
}
