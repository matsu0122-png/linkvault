import { useState, type FormEvent } from 'react'
import Button from '../../../components/ui/Button'
import Input from '../../../components/ui/Input'

export type CollectionFormValues = {
  name: string
  description: string
}

const emptyValues: CollectionFormValues = { name: '', description: '' }

type Props = {
  initialValues?: CollectionFormValues
  submitLabel: string
  onSubmit: (values: CollectionFormValues) => Promise<void>
  onCancel?: () => void
}

function CollectionForm({
  initialValues = emptyValues,
  submitLabel,
  onSubmit,
  onCancel,
}: Props) {
  const [values, setValues] = useState(initialValues)
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)

  function update<K extends keyof CollectionFormValues>(
    key: K,
    value: CollectionFormValues[K],
  ) {
    setValues((prev) => ({ ...prev, [key]: value }))
  }

  async function handleSubmit(e: FormEvent) {
    e.preventDefault()
    setSubmitting(true)
    setError(null)

    try {
      await onSubmit(values)
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err))
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <form onSubmit={handleSubmit} className="flex flex-col gap-1.5">
      <Input
        type="text"
        placeholder="Collection名"
        required
        value={values.name}
        onChange={(e) => update('name', e.target.value)}
        className="text-sm"
      />
      <Input
        type="text"
        placeholder="説明（任意）"
        value={values.description}
        onChange={(e) => update('description', e.target.value)}
        className="text-sm"
      />
      <div className="flex gap-2">
        <Button type="submit" size="sm" disabled={submitting}>
          {submitLabel}
        </Button>
        {onCancel && (
          <Button
            type="button"
            variant="ghost"
            size="sm"
            onClick={onCancel}
            disabled={submitting}
          >
            キャンセル
          </Button>
        )}
      </div>
      {error && (
        <p role="alert" className="text-xs text-rose-600">
          エラー: {error}
        </p>
      )}
    </form>
  )
}

export default CollectionForm
