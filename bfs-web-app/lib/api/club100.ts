import { apiRequest } from "../api"

export interface Club100Person {
  id: string
  firstName: string
  lastName: string
  remaining: number
  max: number
}

export interface Club100Remaining {
  elvantoPersonId: string
  remaining: number
  max: number
}

export async function listClub100People(token: string): Promise<Club100Person[]> {
  const response = await apiRequest<{ items: Club100Person[] }>("/v1/club100/people", {
    headers: {
      Authorization: `Bearer ${token}`,
    },
  })
  return response.items
}

export async function getClub100Remaining(token: string, personId: string): Promise<Club100Remaining> {
  return apiRequest<Club100Remaining>(`/v1/club100/remaining/${personId}`, {
    headers: {
      Authorization: `Bearer ${token}`,
    },
  })
}
