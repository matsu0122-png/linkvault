import { useState } from 'react'
import Button from '../components/ui/Button'
import Input from '../components/ui/Input'
import { syncLinkCollections } from '../features/collections/api'
import type { Collection } from '../features/collections/types'
import BulkLinkForm from '../features/links/components/BulkLinkForm'
import LinkForm from '../features/links/components/LinkForm'
import LinkList from '../features/links/components/LinkList'
import Pagination from '../features/links/components/Pagination'
import { useLinks } from '../features/links/hooks/useLinks'
import type { SortOption } from '../features/links/types'

type Mode = 'none' | 'single' | 'bulk'

const sortLabels: Record<SortOption, string> = {
  created_at_desc: '登録日時（新しい順）',
  created_at_asc: '登録日時（古い順）',
  updated_at_desc: '更新日時（新しい順）',
  updated_at_asc: '更新日時（古い順）',
  title_asc: 'タイトル（あいうえお順）',
  title_desc: 'タイトル（逆順）',
}

type Props = {
  activeCollection: Collection | null
  collections: Collection[]
  onCollectionsChanged: () => void
}

function LinksPage({
  activeCollection,
  collections,
  onCollectionsChanged,
}: Props) {
  const {
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
    goToPage,
    refetch,
    addLink,
    bulkAddLinks,
    checking,
    checkLinks,
    editLink,
    removeLink,
  } = useLinks(activeCollection?.id ?? null)
  const [mode, setMode] = useState<Mode>('none')

  // 所属Collectionが変わったリンクは、現在のCollection絞り込みに
  // 一致しなくなっている可能性があるため一覧を再取得する。サイドバーの
  // 件数(link_count)も併せて更新する。
  async function syncCollectionsAndRefetch(
    linkId: number,
    previousIds: number[],
    nextIds: number[],
  ) {
    if (previousIds.length === 0 && nextIds.length === 0) {
      return
    }
    await syncLinkCollections(linkId, previousIds, nextIds)
    onCollectionsChanged()
    await refetch()
  }

  return (
    <div className="mx-auto flex w-full max-w-2xl flex-col gap-6 px-4 py-16">
      <header className="flex items-baseline justify-between">
        <div>
          <h1 className="font-display text-2xl font-medium text-stone-800">
            {activeCollection ? activeCollection.name : 'すべてのリンク'}
          </h1>
          {activeCollection?.description && (
            <p className="mt-0.5 text-sm text-stone-500">
              {activeCollection.description}
            </p>
          )}
        </div>
        {mode === 'none' && (
          <div className="flex items-center gap-3">
            <Button type="button" size="sm" onClick={() => setMode('single')}>
              + 新しいリンク
            </Button>
            <button
              type="button"
              onClick={() => setMode('bulk')}
              className="text-xs text-stone-500 hover:text-teal-600 hover:underline"
            >
              まとめて登録
            </button>
          </div>
        )}
      </header>

      <div className="flex flex-wrap items-center gap-2">
        <Input
          type="search"
          placeholder="検索"
          aria-label="検索"
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          className="flex-1"
        />
        <select
          value={sort}
          onChange={(e) => setSort(e.target.value as SortOption)}
          aria-label="並び替え"
          className="rounded-xl border border-stone-300 bg-white px-2.5 py-2 text-sm text-stone-600 focus:border-teal-500 focus:ring-1 focus:ring-teal-500 focus:outline-none"
        >
          {Object.entries(sortLabels).map(([value, label]) => (
            <option key={value} value={value}>
              {label}
            </option>
          ))}
        </select>
        {activeTag && (
          <button
            type="button"
            onClick={() => setActiveTag(null)}
            className="flex items-center gap-1 rounded-full bg-teal-600 px-3 py-1 text-xs font-medium text-white hover:bg-teal-700"
          >
            #{activeTag} ✕
          </button>
        )}
        <button
          type="button"
          onClick={() => checkLinks()}
          disabled={checking}
          className="shrink-0 text-xs text-stone-500 hover:text-teal-600 hover:underline disabled:cursor-not-allowed disabled:opacity-50"
        >
          {checking ? 'チェック中...' : 'リンク切れをチェック'}
        </button>
      </div>

      {mode === 'single' && (
        <LinkForm
          submitLabel="追加"
          collections={collections}
          onSubmit={async (values) => {
            const created = await addLink(values)
            await syncCollectionsAndRefetch(created.id, [], values.collectionIds)
          }}
          onCancel={() => setMode('none')}
        />
      )}

      {mode === 'bulk' && (
        <BulkLinkForm onSubmit={bulkAddLinks} onCancel={() => setMode('none')} />
      )}

      {error && (
        <p
          role="alert"
          className="rounded-xl border border-rose-200 bg-rose-50 px-3 py-2 text-sm text-rose-700"
        >
          エラー: {error}
        </p>
      )}

      <div
        className={loading ? 'opacity-50 transition' : 'transition'}
        aria-busy={loading}
      >
        <LinkList
          links={links}
          loading={loading}
          onEdit={editLink}
          onDelete={removeLink}
          onTagClick={setActiveTag}
          collections={collections}
          onCollectionsSync={syncCollectionsAndRefetch}
        />
      </div>

      <Pagination
        pagination={pagination}
        onPageChange={goToPage}
        disabled={loading}
      />
    </div>
  )
}

export default LinksPage
