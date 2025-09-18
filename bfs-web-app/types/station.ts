export interface Station {
  id: string;
  name: string;
  createdAt: string; // ISO date
  updatedAt: string; // ISO date
}

export type StationRequestStatus = "pending" | "approved" | "rejected";

export interface StationRequest {
  id: string;
  name: string;
  model: string;
  os: string;
  status: StationRequestStatus;
  decidedBy: string | null;
  decidedAt: string | null; // ISO date
  createdAt: string; // ISO date
  expiresAt: string; // ISO date
}

export interface StationProduct {
  id: string;
  stationId: string;
  productId: string;
}