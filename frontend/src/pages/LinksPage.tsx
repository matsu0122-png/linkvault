import { useState } from 'react'
import Button from '../components/ui/Button'
import Input from '../components/ui/Input'
import LinkForm from '../features/links/components/LinkForm'
import LinkList from '../features/links/components/LinkList'
import { useLinks } from '../features/links/hooks/useLinks'

function LinksPage() {
  const {
    links,
    error,
    query,
    setQuery,
    activeTag,
    setActiveTag,
    addLink,
    editLink,
    removeLink,
  } = useLinks()
  const [adding, setAdding] = useState(false)

  return (
    <div className="mx-auto flex w-full max-w-2xl flex-col gap-6 px-4 py-16">
      <header className="flex items-baseline justify-between">
        <h1 className="font-display text-2xl font-medium text-stone-800">
          LinkVault
        </h1>
        {!adding && (
          <Button type="button" size="sm" onClick={() => setAdding(true)}>
            + 新しいリンク
          </Button>
        )}
      </header>

      <div className="flex flex-wrap items-center gap-2">
        <Input
          type="search"
          placeholder="検索"
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          className="flex-1"
        />
        {activeTag && (
          <button
            type="button"
            onClick={() => setActiveTag(null)}
            className="flex items-center gap-1 rounded-full bg-teal-600 px-3 py-1 text-xs font-medium text-white hover:bg-teal-700"
          >
            #{activeTag} ✕
          </button>
        )}
      </div>

      {adding && (
        <LinkForm
          submitLabel="追加"
          onSubmit={async (values) => {
            await addLink(values)
          }}
          onCancel={() => setAdding(false)}
        />
      )}

      {error && (
        <p
          role="alert"
          className="rounded-xl border border-rose-200 bg-rose-50 px-3 py-2 text-sm text-rose-700"
        >
          エラー: {error}
        </p>
      )}

      <LinkList
        links={links}
        onEdit={editLink}
        onDelete={removeLink}
        onTagClick={setActiveTag}
      />
    </div>
  )
}

export default LinksPage
