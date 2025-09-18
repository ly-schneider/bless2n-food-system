export type AdminInviteStatus = "pending" | "accepted" | "expired" | "revoked";

export interface AdminInvite {
  id: string;
  invitedBy: string;
  inviteeEmail: string;
  expiresAt: string; // ISO date
  status: AdminInviteStatus;
  usedAt: string | null; // ISO date
  createdAt: string; // ISO date
}