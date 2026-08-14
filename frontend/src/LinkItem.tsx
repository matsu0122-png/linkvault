import { useState, type FormEvent } from 'react'
import { deleteLink, updateLink } from './api/links'
import type { Link } from './types'

type Props = {
  link: Link
  onUpdated: (link: Link) => void
  onDeleted: (id: number) => void
}

function LinkItem({ link, onUpdated, onDeleted }: Props) {
  const [editing, setEditing] = useState(false)
  const [url, setUrl] = useState(link.url)
  const [title, setTitle] = useState(link.title)
  const [memo, setMemo] = useState(link.memo)
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)

  async function handleSave(e: FormEvent) {
    e.preventDefault()
    setSubmitting(true)
    setError(null)

    try {
      const updated = await updateLink(link.id, { url, title, memo })
      onUpdated(updated)
      setEditing(false)
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err))
    } finally {
      setSubmitting(false)
    }
  }

  async function handleDelete() {
    setSubmitting(true)
    setError(null)

    try {
      await deleteLink(link.id)
      onDeleted(link.id)
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err))
      setSubmitting(false)
    }
  }

  function handleCancel() {
    setUrl(link.url)
    setTitle(link.title)
    setMemo(link.memo)
    setEditing(false)
  }

  if (editing) {
    return (
      <li>
        <form onSubmit={handleSave}>
          <input
            type="url"
            value={url}
            onChange={(e) => setUrl(e.target.value)}
            required
          />
          <input
            type="text"
            value={title}
            onChange={(e) => setTitle(e.target.value)}
          />
          <input
            type="text"
            value={memo}
            onChange={(e) => setMemo(e.target.value)}
          />
          <button type="submit" disabled={submitting}>
            保存
          </button>
          <button type="button" onClick={handleCancel} disabled={submitting}>
            キャンセル
          </button>
        </form>
        {error && <p role="alert">エラー: {error}</p>}
      </li>
    )
  }

  return (
    <li>
      <a href={link.url} target="_blank" rel="noreferrer">
        {link.title}
      </a>
      {link.memo && <p>{link.memo}</p>}
      <button type="button" onClick={() => setEditing(true)}>
        編集
      </button>
      <button type="button" onClick={handleDelete} disabled={submitting}>
        削除
      </button>
      {error && <p role="alert">エラー: {error}</p>}
    </li>
  )
}

export default LinkItem
