import type { Link } from './types'

const API_BASE_URL =
  import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:8080'

export type LinkFilter = {
  query?: string
  tag?: string
}

export async function fetchLinks(filter: LinkFilter = {}): Promise<Link[]> {
  const params = new URLSearchParams()
  if (filter.query) params.set('q', filter.query)
  if (filter.tag) params.set('tag', filter.tag)

  const res = await fetch(`${API_BASE_URL}/api/links?${params}`)
  if (!res.ok) {
    throw new Error(`failed to fetch links: ${res.status}`)
  }
  return res.json()
}

export type CreateLinkInput = {
  url: string
  title: string
  memo: string
  tags: string[]
}

export async function createLink(input: CreateLinkInput): Promise<Link> {
  const res = await fetch(`${API_BASE_URL}/api/links`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(input),
  })
  if (!res.ok) {
    throw new Error(`failed to create link: ${res.status}`)
  }
  return res.json()
}

export type UpdateLinkInput = CreateLinkInput

export async function updateLink(
  id: number,
  input: UpdateLinkInput,
): Promise<Link> {
  const res = await fetch(`${API_BASE_URL}/api/links/${id}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(input),
  })
  if (!res.ok) {
    throw new Error(`failed to update link: ${res.status}`)
  }
  return res.json()
}

export type BulkCreateInput = {
  urls: string[]
  tags: string[]
}

export type BulkCreateFailure = {
  url: string
  error: string
}

export type BulkCreateResult = {
  created: Link[]
  failed: BulkCreateFailure[]
}

export async function bulkCreateLinks(
  input: BulkCreateInput,
): Promise<BulkCreateResult> {
  const res = await fetch(`${API_BASE_URL}/api/links/bulk`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(input),
  })
  if (!res.ok) {
    throw new Error(`failed to bulk create links: ${res.status}`)
  }
  return res.json()
}

export async function deleteLink(id: number): Promise<void> {
  const res = await fetch(`${API_BASE_URL}/api/links/${id}`, {
    method: 'DELETE',
  })
  if (!res.ok) {
    throw new Error(`failed to delete link: ${res.status}`)
  }
}
