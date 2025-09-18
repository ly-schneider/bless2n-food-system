export interface OTPToken {
  id: string;
  userId: string;
  tokenHash: string;
  createdAt: string; // ISO date
  usedAt: string | null; // ISO date
  attempts: number;
  expiresAt: string; // ISO date
}

export interface RefreshToken {
  id: string;
  userId: string;
  clientId: string;
  tokenHash: string;
  issuedAt: string; // ISO date
  lastUsedAt: string; // ISO date
  expiresAt: string; // ISO date
  isRevoked: boolean;
  revokedReason: string | null;
  familyId: string;
}