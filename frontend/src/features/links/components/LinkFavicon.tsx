import { useState } from 'react'
import type { Link } from '../types'

type Props = {
  link: Link
}

function initial(link: Link): string {
  const source = (link.title || link.url).trim()
  return source ? source.charAt(0).toUpperCase() : '?'
}

// favicon_urlが無い、またはURLはあっても実際には読み込めなかった場合に、
// 頭文字だけのバッジへ差し替える。onErrorで失敗を検知した後は二度と
// <img>へ戻さない（壊れたURLへ何度もリクエストし直さないため）。
function LinkFavicon({ link }: Props) {
  const [failed, setFailed] = useState(false)

  if (!link.favicon_url || failed) {
    return (
      <span
        aria-hidden="true"
        className="flex h-3.5 w-3.5 shrink-0 translate-y-0.5 items-center justify-center rounded-full bg-stone-200 text-[8px] font-medium leading-none text-stone-500"
      >
        {initial(link)}
      </span>
    )
  }

  return (
    <img
      src={link.favicon_url}
      alt=""
      className="h-3.5 w-3.5 shrink-0 translate-y-0.5 rounded-sm"
      onError={() => setFailed(true)}
    />
  )
}

export default LinkFavicon
