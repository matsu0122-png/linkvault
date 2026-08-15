import type { Collection, CreateCollectionInput, UpdateCollectionInput } from './types'

const API_BASE_URL =
  import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:8080'

export async function fetchCollections(linkId?: number): Promise<Collection[]> {
  const params = new URLSearchParams()
  if (linkId) params.set('link_id', String(linkId))

  const res = await fetch(`${API_BASE_URL}/api/collections?${params}`)
  if (!res.ok) {
    throw new Error(`failed to fetch collections: ${res.status}`)
  }
  return res.json()
}

export async function createCollection(
  input: CreateCollectionInput,
): Promise<Collection> {
  const res = await fetch(`${API_BASE_URL}/api/collections`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(input),
  })
  if (!res.ok) {
    throw new Error(`failed to create collection: ${res.status}`)
  }
  return res.json()
}

export async function updateCollection(
  id: number,
  input: UpdateCollectionInput,
): Promise<Collection> {
  const res = await fetch(`${API_BASE_URL}/api/collections/${id}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(input),
  })
  if (!res.ok) {
    throw new Error(`failed to update collection: ${res.status}`)
  }
  return res.json()
}

export async function deleteCollection(id: number): Promise<void> {
  const res = await fetch(`${API_BASE_URL}/api/collections/${id}`, {
    method: 'DELETE',
  })
  if (!res.ok) {
    throw new Error(`failed to delete collection: ${res.status}`)
  }
}

export async function addLinkToCollection(
  collectionId: number,
  linkId: number,
): Promise<void> {
  const res = await fetch(
    `${API_BASE_URL}/api/collections/${collectionId}/links`,
    {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ link_id: linkId }),
    },
  )
  if (!res.ok) {
    throw new Error(`failed to add link to collection: ${res.status}`)
  }
}

export async function removeLinkFromCollection(
  collectionId: number,
  linkId: number,
): Promise<void> {
  const res = await fetch(
    `${API_BASE_URL}/api/collections/${collectionId}/links/${linkId}`,
    { method: 'DELETE' },
  )
  if (!res.ok) {
    throw new Error(`failed to remove link from collection: ${res.status}`)
  }
}

// syncLinkCollections reconciles a link's collection membership from
// previousIds to nextIds by adding/removing only what changed, so a Link
// create/edit form can submit "the full set the user checked" without the
// caller having to diff it manually.
export async function syncLinkCollections(
  linkId: number,
  previousIds: number[],
  nextIds: number[],
): Promise<void> {
  const toAdd = nextIds.filter((id) => !previousIds.includes(id))
  const toRemove = previousIds.filter((id) => !nextIds.includes(id))

  await Promise.all([
    ...toAdd.map((id) => addLinkToCollection(id, linkId)),
    ...toRemove.map((id) => removeLinkFromCollection(id, linkId)),
  ])
}
