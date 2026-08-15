import type { Collection } from '../types'
import { buildCollectionTree, flattenCollectionTree } from '../tree'

type Props = {
  collections: Collection[]
  selectedIds: number[]
  onChange: (ids: number[]) => void
}

const INDENT_PX = 14

function CollectionSelect({ collections, selectedIds, onChange }: Props) {
  function toggle(id: number) {
    if (selectedIds.includes(id)) {
      onChange(selectedIds.filter((i) => i !== id))
    } else {
      onChange([...selectedIds, id])
    }
  }

  if (collections.length === 0) {
    return <p className="text-xs text-stone-400">Collectionがまだありません</p>
  }

  const flat = flattenCollectionTree(buildCollectionTree(collections))

  return (
    <div className="flex flex-col gap-0.5 rounded-xl border border-stone-300 bg-white p-1.5">
      {flat.map(({ collection: c, depth }) => (
        <label
          key={c.id}
          className="flex cursor-pointer items-center gap-2 rounded-lg py-1 pr-2 text-sm text-stone-700 hover:bg-stone-50"
          style={{ paddingLeft: `${depth * INDENT_PX + 8}px` }}
        >
          <input
            type="checkbox"
            checked={selectedIds.includes(c.id)}
            onChange={() => toggle(c.id)}
            className="accent-teal-600"
          />
          {c.name}
        </label>
      ))}
    </div>
  )
}

export default CollectionSelect
