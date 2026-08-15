import { useState, type FormEvent } from 'react'
import Button from '../../../components/ui/Button'
import Input from '../../../components/ui/Input'
import TagInput from '../../../components/ui/TagInput'
import CollectionSelect from '../../collections/components/CollectionSelect'
import type { Collection } from '../../collections/types'

export type LinkFormValues = {
  url: string
  title: string
  memo: string
  tags: string[]
  collectionIds: number[]
}

const emptyValues: LinkFormValues = {
  url: '',
  title: '',
  memo: '',
  tags: [],
  collectionIds: [],
}

type Props = {
  initialValues?: LinkFormValues
  submitLabel: string
  onSubmit: (values: LinkFormValues) => Promise<void>
  onCancel?: () => void
  collections: Collection[]
}

function LinkForm({
  initialValues = emptyValues,
  submitLabel,
  onSubmit,
  onCancel,
  collections,
}: Props) {
  const [values, setValues] = useState(initialValues)
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)

  function update<K extends keyof LinkFormValues>(key: K, value: LinkFormValues[K]) {
    setValues((prev) => ({ ...prev, [key]: value }))
  }

  async function handleSubmit(e: FormEvent) {
    e.preventDefault()
    setSubmitting(true)
    setError(null)

    try {
      await onSubmit(values)
      setValues(emptyValues)
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err))
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <form
      onSubmit={handleSubmit}
      className="flex flex-col gap-2 sm:flex-row sm:items-start"
    >
      <div className="grid flex-1 gap-2 sm:grid-cols-4">
        <Input
          type="url"
          placeholder="URL"
          required
          value={values.url}
          onChange={(e) => update('url', e.target.value)}
          className="sm:col-span-2"
        />
        <Input
          type="text"
          placeholder="タイトル"
          value={values.title}
          onChange={(e) => update('title', e.target.value)}
        />
        <Input
          type="text"
          placeholder="メモ"
          value={values.memo}
          onChange={(e) => update('memo', e.target.value)}
        />
        <TagInput
          value={values.tags}
          onChange={(tags) => update('tags', tags)}
          className="sm:col-span-4"
        />
        <div className="sm:col-span-4">
          <p className="mb-1 text-xs text-stone-500">Collection</p>
          <CollectionSelect
            collections={collections}
            selectedIds={values.collectionIds}
            onChange={(ids) => update('collectionIds', ids)}
          />
        </div>
      </div>
      <div className="flex gap-2">
        <Button type="submit" disabled={submitting}>
          {submitLabel}
        </Button>
        {onCancel && (
          <Button
            type="button"
            variant="ghost"
            onClick={onCancel}
            disabled={submitting}
          >
            キャンセル
          </Button>
        )}
      </div>
      {error && (
        <p role="alert" className="text-sm text-rose-600 sm:basis-full">
          エラー: {error}
        </p>
      )}
    </form>
  )
}

export default LinkForm
