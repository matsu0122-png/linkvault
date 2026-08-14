import type { Link } from '../types'

const API_BASE_URL =
  import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:8080'

export async function fetchLinks(): Promise<Link[]> {
  const res = await fetch(`${API_BASE_URL}/api/links`)
  if (!res.ok) {
    throw new Error(`failed to fetch links: ${res.status}`)
  }
  return res.json()
}
