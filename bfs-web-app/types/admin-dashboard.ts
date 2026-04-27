export type DashboardStatusLevel = "green" | "yellow" | "red"

export interface SystemStatusChip {
  key: string
  label: string
  status: DashboardStatusLevel
  summary: string
}

export interface StationOverviewCard {
  id: string
  name: string
  status: DashboardStatusLevel
  openOrders: number
  backlog: number
  medianThroughputMinutes: number
  recentProductTitle?: string | null
  recentProductQuantity: number
}

export interface AdminOpsOverview {
  overallStatus: DashboardStatusLevel
  statusChips: SystemStatusChip[]
  stations: StationOverviewCard[]
}

export interface DashboardTopProductWindowItem {
  title: string
  quantity: number
  revenueCents: number
}

export interface DashboardSeriesPoint {
  label: string
  value: number
}

export interface StationQueueMetric {
  orderId: string
  createdAt: string
  ageMinutes: number
  pendingItems: number
  pendingQuantity: number
  titles: string[]
}

export interface StationDetailSummary {
  station: StationOverviewCard
  queue: StationQueueMetric[]
  topProducts: DashboardTopProductWindowItem[]
  throughputByHour: DashboardSeriesPoint[]
}
