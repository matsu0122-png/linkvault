import type { Link } from '../types'
import LinkRow from './LinkRow'
import type { LinkFormValues } from './LinkForm'

type Props = {
  links: Link[]
  onEdit: (id: number, values: LinkFormValues) => Promise<void>
  onDelete: (id: number) => Promise<void>
  onTagClick: (tag: string) => void
}

function LinkList({ links, onEdit, onDelete, onTagClick }: Props) {
  if (links.length === 0) {
    return (
      <p className="rounded-2xl border border-dashed border-stone-300 bg-white/50 py-10 text-center text-sm text-stone-400">
        リンクがまだありません
      </p>
    )
  }

  return (
    <ul className="flex flex-col gap-2 rounded-2xl border border-stone-200 bg-white p-2 shadow-sm shadow-stone-200/50">
      {links.map((link) => (
        <LinkRow
          key={link.id}
          link={link}
          onEdit={onEdit}
          onDelete={onDelete}
          onTagClick={onTagClick}
        />
      ))}
    </ul>
  )
}

export default LinkList
