export const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL as string;

if (!API_BASE_URL) {
  throw new Error('NEXT_PUBLIC_API_BASE_URL environment variable is required');
}

export interface ApiError {
  status: number;
  message: string;
}

export function createApiError(status: number, message: string): ApiError {
  return { status, message };
}

export async function apiRequest<T>(
  endpoint: string,
  options: RequestInit = {}
): Promise<T> {
  const url = `${API_BASE_URL}${endpoint}`;
  
  const response = await fetch(url, {
    headers: {
      'Content-Type': 'application/json',
      ...options.headers,
    },
    ...options,
  });

  if (!response.ok) {
    const errorData = await response.json().catch(() => ({})) as { message?: string; detail?: string };
    const error = createApiError(
      response.status,
      errorData.message || errorData.detail || `HTTP ${response.status}: ${response.statusText}`
    );
    throw error;
  }

  return response.json() as Promise<T>;
}

export { };
