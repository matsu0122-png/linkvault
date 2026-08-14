import { useState, type FormEvent } from 'react'
import Button from '../../../components/ui/Button'
import Input from '../../../components/ui/Input'

export type LinkFormValues = {
  url: string
  title: string
  memo: string
}

const emptyValues: LinkFormValues = { url: '', title: '', memo: '' }

type Props = {
  initialValues?: LinkFormValues
  submitLabel: string
  onSubmit: (values: LinkFormValues) => Promise<void>
  onCancel?: () => void
}

function LinkForm({
  initialValues = emptyValues,
  submitLabel,
  onSubmit,
  onCancel,
}: Props) {
  const [values, setValues] = useState(initialValues)
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)

  function update<K extends keyof LinkFormValues>(key: K, value: string) {
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
