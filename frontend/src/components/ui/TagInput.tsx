import { useState, type KeyboardEvent } from 'react'

type Props = {
  value: string[]
  onChange: (tags: string[]) => void
  placeholder?: string
  className?: string
}

function TagInput({
  value,
  onChange,
  placeholder = 'タグを追加してEnter',
  className = '',
}: Props) {
  const [draft, setDraft] = useState('')

  function commit() {
    const tag = draft.trim()
    if (tag && !value.includes(tag)) {
      onChange([...value, tag])
    }
    setDraft('')
  }

  function handleKeyDown(e: KeyboardEvent<HTMLInputElement>) {
    if (e.key === 'Enter' || e.key === ',') {
      e.preventDefault()
      commit()
    } else if (e.key === 'Backspace' && draft === '' && value.length > 0) {
      onChange(value.slice(0, -1))
    }
  }

  function removeTag(tag: string) {
    onChange(value.filter((t) => t !== tag))
  }

  return (
    <div
      className={`flex flex-wrap items-center gap-1.5 rounded-xl border border-stone-300 bg-white px-2 py-1.5 focus-within:border-teal-500 focus-within:ring-1 focus-within:ring-teal-500 ${className}`}
    >
      {value.map((tag) => (
        <span
          key={tag}
          className="flex items-center gap-1 rounded-full bg-teal-50 px-2 py-0.5 text-xs text-teal-700"
        >
          {tag}
          <button
            type="button"
            onClick={() => removeTag(tag)}
            className="text-teal-500 hover:text-teal-800"
            aria-label={`${tag}を削除`}
          >
            ×
          </button>
        </span>
      ))}
      <input
        type="text"
        value={draft}
        onChange={(e) => setDraft(e.target.value)}
        onKeyDown={handleKeyDown}
        onBlur={commit}
        placeholder={value.length === 0 ? placeholder : ''}
        className="min-w-24 flex-1 bg-transparent text-sm text-stone-800 placeholder:text-stone-400 focus:outline-none"
      />
    </div>
  )
}

export default TagInput
