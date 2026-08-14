import { useCallback, useEffect, useState } from 'react'
import {
  bulkCreateLinks,
  checkLinks as checkLinksRequest,
  createLink,
  deleteLink,
  fetchLinks,
  updateLink,
  type BulkCreateInput,
  type BulkCreateResult,
  type CreateLinkInput,
  type UpdateLinkInput,
} from '../api'
import type { Link } from '../types'

function toMessage(err: unknown): string {
  return err instanceof Error ? err.message : String(err)
}

export function useLinks() {
  const [links, setLinks] = useState<Link[]>([])
  const [error, setError] = useState<string | null>(null)
  const [query, setQuery] = useState('')
  const [activeTag, setActiveTag] = useState<string | null>(null)
  const [checking, setChecking] = useState(false)

  useEffect(() => {
    const timer = setTimeout(() => {
      fetchLinks({ query, tag: activeTag ?? undefined })
        .then(setLinks)
        .catch((err: unknown) => setError(toMessage(err)))
    }, 300)

    return () => clearTimeout(timer)
  }, [query, activeTag])

  const addLink = useCallback(async (input: CreateLinkInput) => {
    const created = await createLink(input)
    setLinks((prev) => [...prev, created])
  }, [])

  const bulkAddLinks = useCallback(
    async (input: BulkCreateInput): Promise<BulkCreateResult> => {
      const result = await bulkCreateLinks(input)
      setLinks((prev) => [...prev, ...result.created])
      return result
    },
    [],
  )

  const checkLinks = useCallback(async () => {
    setChecking(true)
    try {
      await checkLinksRequest()
      // checkLinksRequest updates every link regardless of the current
      // filter, so refetch through the filtered endpoint to stay consistent
      // with what's on screen.
      const refreshed = await fetchLinks({ query, tag: activeTag ?? undefined })
      setLinks(refreshed)
    } finally {
      setChecking(false)
    }
  }, [query, activeTag])

  const editLink = useCallback(async (id: number, input: UpdateLinkInput) => {
    const updated = await updateLink(id, input)
    setLinks((prev) => prev.map((l) => (l.id === updated.id ? updated : l)))
  }, [])

  const removeLink = useCallback(async (id: number) => {
    await deleteLink(id)
    setLinks((prev) => prev.filter((l) => l.id !== id))
  }, [])

  return {
    links,
    error,
    query,
    setQuery,
    activeTag,
    setActiveTag,
    addLink,
    bulkAddLinks,
    checking,
    checkLinks,
    editLink,
    removeLink,
  }
}
