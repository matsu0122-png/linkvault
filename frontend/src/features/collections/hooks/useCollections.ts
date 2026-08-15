import { useCallback, useEffect, useState } from 'react'
import {
  createCollection,
  deleteCollection,
  fetchCollections,
  updateCollection,
} from '../api'
import type {
  Collection,
  CreateCollectionInput,
  UpdateCollectionInput,
} from '../types'

function toMessage(err: unknown): string {
  return err instanceof Error ? err.message : String(err)
}

function byName(a: Collection, b: Collection): number {
  return a.name.localeCompare(b.name)
}

export function useCollections() {
  const [collections, setCollections] = useState<Collection[]>([])
  const [error, setError] = useState<string | null>(null)

  const refresh = useCallback(async () => {
    try {
      const list = await fetchCollections()
      setCollections(list)
    } catch (err) {
      setError(toMessage(err))
    }
  }, [])

  useEffect(() => {
    refresh()
  }, [refresh])

  const addCollection = useCallback(async (input: CreateCollectionInput) => {
    const created = await createCollection(input)
    setCollections((prev) => [...prev, created].sort(byName))
    return created
  }, [])

  const editCollection = useCallback(
    async (id: number, input: UpdateCollectionInput) => {
      const updated = await updateCollection(id, input)
      setCollections((prev) =>
        prev.map((c) => (c.id === updated.id ? updated : c)).sort(byName),
      )
      return updated
    },
    [],
  )

  const removeCollection = useCallback(
    async (id: number) => {
      await deleteCollection(id)
      // 削除したCollectionに子がいた場合、子は最上位へ昇格する
      // (parent_idがnullに変わる)。ローカルでの部分更新だとその変化を
      // 反映できないため、素直に一覧を取り直す。
      await refresh()
    },
    [refresh],
  )

  return {
    collections,
    error,
    refresh,
    addCollection,
    editCollection,
    removeCollection,
  }
}
