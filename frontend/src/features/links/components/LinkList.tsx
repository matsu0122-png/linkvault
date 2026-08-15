import type { Collection } from '../../collections/types'
import type { Link } from '../types'
import LinkRow from './LinkRow'
import type { LinkFormValues } from './LinkForm'

type Props = {
  links: Link[]
  loading: boolean
  onEdit: (id: number, values: LinkFormValues) => Promise<Link>
  onDelete: (id: number) => Promise<void>
  onTagClick: (tag: string) => void
  collections: Collection[]
  onCollectionsSync: (
    linkId: number,
    previousIds: number[],
    nextIds: number[],
  ) => Promise<void>
}

function LinkList({
  links,
  loading,
  onEdit,
  onDelete,
  onTagClick,
  collections,
  onCollectionsSync,
}: Props) {
  if (links.length === 0) {
    // 読み込み中は「リンクがまだありません」を出さない。取得前の一瞬
    // だけこのメッセージが出てすぐ消える、というちらつきを避けるため。
    if (loading) {
      return null
    }

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
          collections={collections}
          onCollectionsSync={onCollectionsSync}
        />
      ))}
    </ul>
  )
}

export default LinkList
