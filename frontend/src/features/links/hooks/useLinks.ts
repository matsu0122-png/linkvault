import { useCallback, useEffect, useState } from 'react'
import {
  createLink,
  deleteLink,
  fetchLinks,
  updateLink,
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
    editLink,
    removeLink,
  }
}
