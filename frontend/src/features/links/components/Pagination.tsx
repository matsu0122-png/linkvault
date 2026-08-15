import type { Pagination as PaginationType } from '../types'

type Props = {
  pagination: PaginationType
  onPageChange: (page: number) => void
  disabled?: boolean
}

type PageItem = number | 'ellipsis'

function pageItems(current: number, total: number): PageItem[] {
  const items: PageItem[] = [1]

  const left = Math.max(2, current - 1)
  const right = Math.min(total - 1, current + 1)

  if (left > 2) items.push('ellipsis')
  for (let i = left; i <= right; i++) items.push(i)
  if (right < total - 1) items.push('ellipsis')

  if (total > 1) items.push(total)

  return items
}

function Pagination({ pagination, onPageChange, disabled }: Props) {
  const { page, totalPages } = pagination

  if (totalPages <= 1) {
    return null
  }

  return (
    <nav
      aria-label="ページ送り"
      className="flex items-center justify-center gap-1 pt-2"
    >
      <button
        type="button"
        onClick={() => onPageChange(page - 1)}
        disabled={disabled || page <= 1}
        className="rounded-full px-3 py-1.5 text-sm text-stone-500 transition hover:bg-stone-100 hover:text-stone-700 disabled:cursor-not-allowed disabled:opacity-40 disabled:hover:bg-transparent"
      >
        ← 前へ
      </button>

      {pageItems(page, totalPages).map((item, i) =>
        item === 'ellipsis' ? (
          <span
            key={`ellipsis-${i}`}
            className="px-1.5 text-sm text-stone-400"
          >
            …
          </span>
        ) : (
          <button
            key={item}
            type="button"
            onClick={() => onPageChange(item)}
            disabled={disabled}
            aria-current={item === page ? 'page' : undefined}
            className={`min-w-8 rounded-full px-2.5 py-1.5 text-sm transition disabled:cursor-not-allowed ${
              item === page
                ? 'bg-teal-600 font-medium text-white'
                : 'text-stone-600 hover:bg-stone-100'
            }`}
          >
            {item}
          </button>
        ),
      )}

      <button
        type="button"
        onClick={() => onPageChange(page + 1)}
        disabled={disabled || page >= totalPages}
        className="rounded-full px-3 py-1.5 text-sm text-stone-500 transition hover:bg-stone-100 hover:text-stone-700 disabled:cursor-not-allowed disabled:opacity-40 disabled:hover:bg-transparent"
      >
        次へ →
      </button>
    </nav>
  )
}

export default Pagination
