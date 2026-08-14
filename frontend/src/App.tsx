import { useEffect, useState } from 'react'
import { fetchLinks } from './api/links'
import type { Link } from './types'
import './App.css'

function App() {
  const [links, setLinks] = useState<Link[]>([])
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    fetchLinks()
      .then(setLinks)
      .catch((err: unknown) => {
        setError(err instanceof Error ? err.message : String(err))
      })
  }, [])

  return (
    <section id="center">
      <h1>LinkVault</h1>

      {error && <p role="alert">リンクの取得に失敗しました: {error}</p>}

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
