import { useState, type FormEvent } from 'react'
import Button from '../../../components/ui/Button'
import TagInput from '../../../components/ui/TagInput'
import type { BulkCreateInput, BulkCreateResult } from '../api'

type Props = {
  onSubmit: (input: BulkCreateInput) => Promise<BulkCreateResult>
  onCancel: () => void
}

function BulkLinkForm({ onSubmit, onCancel }: Props) {
  const [urlsText, setUrlsText] = useState('')
  const [tags, setTags] = useState<string[]>([])
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [result, setResult] = useState<BulkCreateResult | null>(null)

  async function handleSubmit(e: FormEvent) {
    e.preventDefault()

    const urls = urlsText
      .split('\n')
      .map((line) => line.trim())
      .filter((line) => line !== '')

    if (urls.length === 0) return

    setSubmitting(true)
    setError(null)
    setResult(null)

    try {
      const res = await onSubmit({ urls, tags })
      setResult(res)
      if (res.failed.length === 0) {
        setUrlsText('')
        setTags([])
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err))
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <form onSubmit={handleSubmit} className="flex flex-col gap-2">
      <textarea
        value={urlsText}
        onChange={(e) => setUrlsText(e.target.value)}
        placeholder={
          'URLを1行に1つずつ貼り付け\nhttps://example.com/a\nhttps://example.com/b'
        }
        rows={5}
        className="rounded-xl border border-stone-300 bg-white px-3 py-2 text-sm text-stone-800 placeholder:text-stone-400 focus:border-teal-500 focus:ring-1 focus:ring-teal-500 focus:outline-none"
      />
      <TagInput
        value={tags}
        onChange={setTags}
        placeholder="タグを追加してEnter（全URL共通）"
      />
      <div className="flex items-center gap-2">
        <Button type="submit" disabled={submitting}>
          追加
        </Button>
        <Button
          type="button"
          variant="ghost"
          onClick={onCancel}
          disabled={submitting}
        >
          キャンセル
        </Button>
      </div>
      {result && (
        <p className="text-sm text-stone-500">
          {result.created.length}件登録しました
          {result.failed.length > 0 && `（${result.failed.length}件失敗）`}
        </p>
      )}
      {error && (
        <p role="alert" className="text-sm text-rose-600">
          エラー: {error}
        </p>
      )}
    </form>
  )
}

export default BulkLinkForm
