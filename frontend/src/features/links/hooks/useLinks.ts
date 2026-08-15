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
import type { Link, Pagination, SortOption } from '../types'

function toMessage(err: unknown): string {
  return err instanceof Error ? err.message : String(err)
}

const emptyPagination: Pagination = { page: 1, limit: 20, total: 0, totalPages: 0 }

export function useLinks(collectionId: number | null) {
  const [links, setLinks] = useState<Link[]>([])
  const [pagination, setPagination] = useState<Pagination>(emptyPagination)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [query, setQuery] = useState('')
  const [debouncedQuery, setDebouncedQuery] = useState('')
  const [activeTag, setActiveTag] = useState<string | null>(null)
  const [sort, setSort] = useState<SortOption>('created_at_desc')
  const [page, setPage] = useState(1)
  const [checking, setChecking] = useState(false)

  useEffect(() => {
    const timer = setTimeout(() => setDebouncedQuery(query), 300)
    return () => clearTimeout(timer)
  }, [query])

  // 検索語・タグ・並び順・Collectionが変わったら1ページ目に戻す。中身が
  // 変わらない (例: すでに1ページ目) 場合はpageのstateが変化しないため、
  // load側のuseEffectはload自体の変化（下記）で発火する。
  useEffect(() => {
    setPage(1)
  }, [debouncedQuery, activeTag, sort, collectionId])

  const load = useCallback(
    async (targetPage: number) => {
      setLoading(true)
      setError(null)
      try {
        const res = await fetchLinks({
          query: debouncedQuery,
          tag: activeTag ?? undefined,
          page: targetPage,
          sort,
          collectionId: collectionId ?? undefined,
        })
        setLinks(res.links)
        setPagination(res.pagination)
      } catch (err) {
        setError(toMessage(err))
      } finally {
        setLoading(false)
      }
    },
    [debouncedQuery, activeTag, sort, collectionId],
  )

  useEffect(() => {
    load(page)
  }, [load, page])

  const goToPage = useCallback((targetPage: number) => {
    setPage(targetPage)
  }, [])

  // refetch re-runs the current page's query as-is. Needed after a link's
  // collection membership changes: that can make it stop matching the
  // current collection filter, and unlike add/delete, editLink's local
  // setLinks update has no way to know that on its own.
  const refetch = useCallback(() => load(page), [load, page])

  const addLink = useCallback(
    async (input: CreateLinkInput) => {
      const created = await createLink(input)
      // 新規リンクは並び順のデフォルト(新しい順)なら常に1ページ目に来る。
      // 既に1ページ目にいる場合はpageのstateが変わらないため明示的に再取得する。
      if (page === 1) {
        await load(1)
      } else {
        setPage(1)
      }
      return created
    },
    [page, load],
  )

  const bulkAddLinks = useCallback(
    async (input: BulkCreateInput): Promise<BulkCreateResult> => {
      const result = await bulkCreateLinks(input)
      if (page === 1) {
        await load(1)
      } else {
        setPage(1)
      }
      return result
    },
    [page, load],
  )

  const checkLinks = useCallback(async () => {
    setChecking(true)
    try {
      await checkLinksRequest()
      // checkLinksRequestは全リンクの状態を更新するだけなので、現在の
      // フィルタ・ページに合わせて改めて取得し直す。
      await load(page)
    } finally {
      setChecking(false)
    }
  }, [page, load])

  const editLink = useCallback(async (id: number, input: UpdateLinkInput) => {
    const updated = await updateLink(id, input)
    setLinks((prev) => prev.map((l) => (l.id === updated.id ? updated : l)))
    return updated
  }, [])

  const removeLink = useCallback(
    async (id: number) => {
      await deleteLink(id)

      const wasLastOnPage = links.length === 1
      if (wasLastOnPage && page > 1) {
        setPage((p) => p - 1)
      } else {
        await load(page)
      }
    },
    [links.length, page, load],
  )

  return {
    links,
    pagination,
    loading,
    error,
    query,
    setQuery,
    activeTag,
    setActiveTag,
    sort,
    setSort,
    page,
    goToPage,
    refetch,
    addLink,
    bulkAddLinks,
    checking,
    checkLinks,
    editLink,
    removeLink,
  }
}
