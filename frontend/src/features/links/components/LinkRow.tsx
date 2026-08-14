import { useState } from 'react'
import Button from '../../../components/ui/Button'
import type { Link } from '../types'
import LinkForm, { type LinkFormValues } from './LinkForm'

type Props = {
  link: Link
  onEdit: (id: number, values: LinkFormValues) => Promise<void>
  onDelete: (id: number) => Promise<void>
  onTagClick: (tag: string) => void
}

function hostname(url: string): string {
  try {
    return new URL(url).hostname
  } catch {
    return url
  }
}

function statusTitle(link: Link): string {
  const checkedAt = link.checked_at
    ? new Date(link.checked_at).toLocaleString()
    : ''
  return link.status === 'ok'
    ? `アクセス可能（最終確認: ${checkedAt}）`
    : `リンク切れの可能性（最終確認: ${checkedAt}）`
}

function LinkRow({ link, onEdit, onDelete, onTagClick }: Props) {
  const [editing, setEditing] = useState(false)
  const [deleting, setDeleting] = useState(false)
  const [error, setError] = useState<string | null>(null)

  async function handleDelete() {
    setDeleting(true)
    setError(null)

    try {
      await onDelete(link.id)
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err))
      setDeleting(false)
    }
  }

  if (editing) {
    return (
      <li className="rounded-xl p-3">
        <LinkForm
          initialValues={{
            url: link.url,
            title: link.title,
            memo: link.memo,
            tags: link.tags,
          }}
          submitLabel="保存"
          onSubmit={async (values) => {
            await onEdit(link.id, values)
            setEditing(false)
          }}
          onCancel={() => setEditing(false)}
        />
      </li>
    )
  }

  return (
    <li className="group flex items-start justify-between gap-4 rounded-xl p-3 transition hover:bg-stone-50">
      <div className="min-w-0">
        <div className="flex flex-wrap items-baseline gap-2">
          <a
            href={link.url}
            target="_blank"
            rel="noreferrer"
            className="inline-flex items-baseline gap-1.5 font-medium text-stone-800 hover:text-teal-600 hover:underline"
          >
            {link.favicon_url && (
              <img
                src={link.favicon_url}
                alt=""
                className="h-3.5 w-3.5 shrink-0 translate-y-0.5 rounded-sm"
                onError={(e) => {
                  e.currentTarget.style.display = 'none'
                }}
              />
            )}
            {link.title || link.url}
          </a>
          <span className="inline-flex items-center gap-1 font-mono text-xs text-stone-400">
            {hostname(link.url)}
            {link.status !== 'unknown' && (
              <span
                title={statusTitle(link)}
                className={`inline-block h-1.5 w-1.5 rounded-full ${
                  link.status === 'ok' ? 'bg-emerald-500' : 'bg-rose-400'
                }`}
              />
            )}
          </span>
        </div>
        {(link.memo || link.description) && (
          <p className="mt-0.5 text-sm text-stone-500">
            {link.memo || link.description}
          </p>
        )}
        {link.tags.length > 0 && (
          <div className="mt-1.5 flex flex-wrap gap-1">
            {link.tags.map((tag) => (
              <button
                key={tag}
                type="button"
                onClick={() => onTagClick(tag)}
                className="rounded-full bg-teal-50 px-2 py-0.5 text-xs text-teal-700 hover:bg-teal-100"
              >
                #{tag}
              </button>
            ))}
          </div>
        )}
        {error && (
          <p role="alert" className="mt-1 text-sm text-rose-600">
            エラー: {error}
          </p>
        )}
      </div>
      <div className="flex shrink-0 gap-1 opacity-0 transition group-hover:opacity-100 group-focus-within:opacity-100">
        <Button
          type="button"
          variant="ghost"
          size="sm"
          onClick={() => setEditing(true)}
        >
          編集
        </Button>
        <Button
          type="button"
          variant="danger"
          size="sm"
          onClick={handleDelete}
          disabled={deleting}
        >
          削除
        </Button>
      </div>
    </li>
  )
}

export default LinkRow
