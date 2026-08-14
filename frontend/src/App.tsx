import { useEffect, useState, type FormEvent } from 'react'
import { createLink, fetchLinks } from './api/links'
import type { Link } from './types'
import './App.css'

function App() {
  const [links, setLinks] = useState<Link[]>([])
  const [error, setError] = useState<string | null>(null)

  const [url, setUrl] = useState('')
  const [title, setTitle] = useState('')
  const [memo, setMemo] = useState('')
  const [submitting, setSubmitting] = useState(false)

  useEffect(() => {
    fetchLinks()
      .then(setLinks)
      .catch((err: unknown) => {
        setError(err instanceof Error ? err.message : String(err))
      })
  }, [])

  async function handleSubmit(e: FormEvent) {
    e.preventDefault()
    setSubmitting(true)
    setError(null)

    try {
      const created = await createLink({ url, title, memo })
      setLinks((prev) => [...prev, created])
      setUrl('')
      setTitle('')
      setMemo('')
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err))
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <section id="center">
      <h1>LinkVault</h1>

      <form onSubmit={handleSubmit}>
        <input
          type="url"
          placeholder="URL"
          value={url}
          onChange={(e) => setUrl(e.target.value)}
          required
        />
        <input
          type="text"
          placeholder="タイトル"
          value={title}
          onChange={(e) => setTitle(e.target.value)}
        />
        <input
          type="text"
          placeholder="メモ"
          value={memo}
          onChange={(e) => setMemo(e.target.value)}
        />
        <button type="submit" disabled={submitting}>
          追加
        </button>
      </form>

      {error && <p role="alert">エラー: {error}</p>}

      {!error && links.length === 0 && <p>リンクがまだありません</p>}

      <ul>
        {links.map((link) => (
          <li key={link.id}>
            <a href={link.url} target="_blank" rel="noreferrer">
              {link.title}
            </a>
            {link.memo && <p>{link.memo}</p>}
          </li>
        ))}
      </ul>
    </section>
  )
}

export default App
