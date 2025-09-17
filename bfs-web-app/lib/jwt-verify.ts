import { jwtVerify } from 'jose';
import { cookies } from 'next/headers';
import { NextRequest } from 'next/server';

const JWKS_URL = process.env.JWKS_URL || process.env.JWT_PUBLIC_KEY_URL;
const JWT_ISSUER = process.env.JWT_ISSUER || 'blessthun-food-api';
const JWT_AUDIENCE = process.env.JWT_AUDIENCE || 'blessthun-food-api';
const ACCESS_TOKEN_COOKIE = 'access_token';

// Clock skew tolerance (60 seconds as per spec)
const CLOCK_TOLERANCE = 60;

export interface JWTPayload {
  sub: string;
  iss: string;
  aud: string | string[];
  exp: number;
  nbf?: number;
  iat: number;
  jti: string;
  role: 'customer' | 'admin' | 'station';
}

export interface VerificationResult {
  valid: boolean;
  payload?: JWTPayload;
  error?: string;
}

// JWKS cache for Ed25519 keys
interface CachedJWKS {
  keys: JWKKey[];
  expiresAt: number;
}

interface JWKKey {
  kty: string;
  crv: string;
  kid: string;
  alg: string;
  use?: string;
  x: string;
}

let jwksCache: CachedJWKS | null = null;
const JWKS_CACHE_TTL = 5 * 60 * 1000; // 5 minutes

/**
 * Fetch JWKS from remote endpoint with caching
 */
async function fetchJWKS(): Promise<JWKKey[]> {
  // Check cache first
  if (jwksCache && Date.now() < jwksCache.expiresAt) {
    return jwksCache.keys;
  }

  if (!JWKS_URL) {
    throw new Error('JWKS_URL not configured');
  }

  try {
    const response = await fetch(JWKS_URL, {
      headers: {
        'Accept': 'application/json',
        'User-Agent': 'bfs-web-app-jwks-client/1.0',
      },
      // Add timeout and retry logic
      next: { 
        revalidate: 300, // 5 minutes cache in Next.js
        tags: ['jwks'] 
      }
    });

    if (!response.ok) {
      throw new Error(`JWKS fetch failed: ${response.status} ${response.statusText}`);
    }

    const jwksResponse: unknown = await response.json();
    
    if (!jwksResponse || typeof jwksResponse !== 'object') {
      throw new Error('Invalid JWKS response format');
    }
    
    const jwks = jwksResponse as { keys?: unknown[] };
    
    if (!jwks.keys || !Array.isArray(jwks.keys)) {
      throw new Error('Invalid JWKS format: missing keys array');
    }

    // Filter for Ed25519 keys only
    const ed25519Keys = jwks.keys.filter((key: unknown): key is JWKKey => {
      if (!key || typeof key !== 'object') return false;
      const keyObj = key as Record<string, unknown>;
      return keyObj.kty === 'OKP' && 
             keyObj.crv === 'Ed25519' && 
             typeof keyObj.kid === 'string' &&
             keyObj.kid.length > 0;
    });

    if (ed25519Keys.length === 0) {
      throw new Error('No Ed25519 keys found in JWKS');
    }

    // Cache the keys
    jwksCache = {
      keys: ed25519Keys,
      expiresAt: Date.now() + JWKS_CACHE_TTL,
    };

    return ed25519Keys;
  } catch (error) {
    // If fetch fails and we have cached keys, use them
    if (jwksCache) {
      console.warn('JWKS fetch failed, using cached keys:', error);
      return jwksCache.keys;
    }
    throw error;
  }
}

/**
 * Find key by kid from JWKS
 */
async function getKey(kid: string): Promise<JWKKey> {
  const keys = await fetchJWKS();
  const key = keys.find(k => k.kid === kid);
  
  if (!key) {
    // Force refresh and try again
    jwksCache = null;
    const refreshedKeys = await fetchJWKS();
    const refreshedKey = refreshedKeys.find(k => k.kid === kid);
    
    if (!refreshedKey) {
      throw new Error(`Key with kid '${kid}' not found in JWKS`);
    }
    
    return refreshedKey;
  }
  
  return key;
}

/**
 * Verify JWT token using JWKS
 */
export async function verifyJWT(token: string): Promise<VerificationResult> {
  try {
    if (!token) {
      return { valid: false, error: 'No token provided' };
    }

    if (!JWKS_URL) {
      return { valid: false, error: 'JWKS URL not configured' };
    }

    // Parse token header to get kid
    const parts = token.split('.');
    if (parts.length !== 3) {
      return { valid: false, error: 'Invalid token format' };
    }

    let header: { alg?: string; typ?: string; kid?: string };
    try {
      const headerPart = parts[0];
      if (!headerPart) {
        return { valid: false, error: 'Missing token header' };
      }
      const parsedHeader = JSON.parse(Buffer.from(headerPart, 'base64url').toString());
      header = parsedHeader as { alg?: string; typ?: string; kid?: string };
    } catch {
      return { valid: false, error: 'Invalid token header' };
    }

    // Validate algorithm
    if (header.alg !== 'EdDSA') {
      return { valid: false, error: `Invalid algorithm: ${header.alg}` };
    }

    // Validate typ if present
    if (header.typ && header.typ !== 'JWT') {
      return { valid: false, error: `Invalid token type: ${header.typ}` };
    }

    // Check for kid
    if (!header.kid) {
      return { valid: false, error: 'Missing kid in token header' };
    }

    // Get public key from JWKS
    const jwk = await getKey(header.kid);
    
    // Import the public key for jose
    const publicKey = await importJWK(jwk);

    // Verify token
    const { payload } = await jwtVerify(token, publicKey, {
      issuer: JWT_ISSUER,
      audience: JWT_AUDIENCE,
      algorithms: ['EdDSA'],
      clockTolerance: CLOCK_TOLERANCE,
    });

    // Additional payload validation
    const now = Math.floor(Date.now() / 1000);
    
    // Check NBF with tolerance
    if (payload.nbf && (now + CLOCK_TOLERANCE) < payload.nbf) {
      return { valid: false, error: 'Token not yet valid' };
    }

    // Check EXP with tolerance  
    if (payload.exp && (now - CLOCK_TOLERANCE) > payload.exp) {
      return { valid: false, error: 'Token expired' };
    }

    // Check IAT
    if (payload.iat && (now + CLOCK_TOLERANCE) < payload.iat) {
      return { valid: false, error: 'Token issued in the future' };
    }

    return {
      valid: true,
      payload: {
        sub: payload.sub as string,
        iss: payload.iss as string,
        aud: payload.aud as string | string[],
        exp: payload.exp as number,
        nbf: payload.nbf as number | undefined,
        iat: payload.iat as number,
        jti: payload.jti as string,
        role: (payload as { role?: string }).role as 'customer' | 'admin' | 'station',
      },
    };

  } catch (error) {
    return {
      valid: false,
      error: error instanceof Error ? error.message : 'Token verification failed',
    };
  }
}

/**
 * Import JWK for jose library
 */
async function importJWK(jwk: JWKKey) {
  const { importJWK } = await import('jose');
  return importJWK(jwk);
}

/**
 * Get and verify token from cookies
 */
export async function verifyTokenFromCookies(): Promise<VerificationResult> {
  const cookieStore = await cookies();
  const token = cookieStore.get(ACCESS_TOKEN_COOKIE)?.value;
  
  if (!token) {
    return { valid: false, error: 'No access token in cookies' };
  }

  return verifyJWT(token);
}

/**
 * Get and verify token from NextRequest
 */
export async function verifyTokenFromRequest(request: NextRequest): Promise<VerificationResult> {
  // First try Authorization header
  const authHeader = request.headers.get('authorization');
  if (authHeader?.startsWith('Bearer ')) {
    const token = authHeader.substring(7);
    return verifyJWT(token);
  }

  // Then try cookie
  const token = request.cookies.get(ACCESS_TOKEN_COOKIE)?.value;
  if (!token) {
    return { valid: false, error: 'No access token found' };
  }

  return verifyJWT(token);
}

/**
 * Middleware helper to require authentication
 */
export async function requireAuth(request: NextRequest): Promise<{ user: JWTPayload } | { error: string }> {
  const result = await verifyTokenFromRequest(request);
  
  if (!result.valid || !result.payload) {
    return { error: result.error || 'Authentication required' };
  }

  return { user: result.payload };
}

/**
 * Middleware helper to require specific role
 */
export async function requireRole(
  request: NextRequest, 
  roles: string | string[]
): Promise<{ user: JWTPayload } | { error: string }> {
  const authResult = await requireAuth(request);
  
  if ('error' in authResult) {
    return authResult;
  }

  const allowedRoles = Array.isArray(roles) ? roles : [roles];
  if (!allowedRoles.includes(authResult.user.role)) {
    return { error: 'Insufficient permissions' };
  }

  return authResult;
}

/**
 * Server action helper for authentication
 */
export async function getServerSession(): Promise<{ user: JWTPayload } | { error: string }> {
  const result = await verifyTokenFromCookies();
  
  if (!result.valid || !result.payload) {
    return { error: result.error || 'Not authenticated' };
  }

  return { user: result.payload };
}