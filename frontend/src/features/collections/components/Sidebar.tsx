import { useState } from 'react'
import type { Collection } from '../types'
import { buildCollectionTree, type CollectionTreeNode } from '../tree'
import CollectionForm, { type CollectionFormValues } from './CollectionForm'

const INDENT_PX = 14

type Props = {
  collections: Collection[]
  activeCollectionId: number | null
  onSelect: (id: number | null) => void
  onCreate: (
    values: CollectionFormValues,
    parentId: number | null,
  ) => Promise<void>
  onEdit: (id: number, values: CollectionFormValues) => Promise<void>
  onDelete: (id: number) => Promise<void>
}

function Sidebar({
  collections,
  activeCollectionId,
  onSelect,
  onCreate,
  onEdit,
  onDelete,
}: Props) {
  // undefined = 何も作成していない, null = 最上位に作成中,
  // number = そのCollectionの子として作成中
  const [creatingParentId, setCreatingParentId] = useState<
    number | null | undefined
  >(undefined)
  const [editingId, setEditingId] = useState<number | null>(null)

  const tree = buildCollectionTree(collections)

  function renderNode(node: CollectionTreeNode, depth: number) {
    const indent = { paddingLeft: `${depth * INDENT_PX + 8}px` }

    if (editingId === node.id) {
      return (
        <li key={node.id} className="py-1" style={indent}>
          <CollectionForm
            initialValues={{ name: node.name, description: node.description }}
            submitLabel="保存"
            onSubmit={async (values) => {
              await onEdit(node.id, values)
              setEditingId(null)
            }}
            onCancel={() => setEditingId(null)}
          />
        </li>
      )
    }

    return (
      <li key={node.id}>
        <div
          className="group relative flex items-center justify-between rounded-lg py-1.5 pr-2 hover:bg-stone-100"
          style={indent}
        >
          <button
            type="button"
            onClick={() => onSelect(node.id)}
            className={`min-w-0 flex-1 truncate text-left text-sm ${
              activeCollectionId === node.id
                ? 'font-medium text-teal-700'
                : 'text-stone-600'
            }`}
          >
            {node.name}
          </button>
          <span
            className="ml-2 shrink-0 text-xs text-stone-400 opacity-100 transition group-hover:opacity-0 group-focus-within:opacity-0 pointer-coarse:opacity-0"
          >
            {node.link_count}
          </span>
          {/* ホバー・キーボードフォーカス・タッチのいずれでも操作できるよう
              opacity(表示/非表示)で切り替える。display:noneにすると
              フォーカス不能になりキーボード・タッチから到達できなくなるため
              使わない。件数表示の上に重ねて幅を変えない。 */}
          <span className="absolute top-1/2 right-2 flex shrink-0 -translate-y-1/2 items-center gap-1.5 rounded-lg bg-stone-100 pl-1 opacity-0 transition group-hover:opacity-100 group-focus-within:opacity-100 pointer-coarse:opacity-100">
            <button
              type="button"
              onClick={() => setCreatingParentId(node.id)}
              className="text-xs text-stone-400 hover:text-teal-600"
              aria-label={`${node.name}の中に新しいCollectionを作成`}
            >
              +
            </button>
            <button
              type="button"
              onClick={() => setEditingId(node.id)}
              className="text-xs text-stone-400 hover:text-teal-600"
            >
              編集
            </button>
            <button
              type="button"
              onClick={() => {
                if (
                  window.confirm(
                    `「${node.name}」を削除しますか？\n中にあるリンクや子Collectionは削除されません。`,
                  )
                ) {
                  onDelete(node.id)
                }
              }}
              className="text-xs text-stone-400 hover:text-rose-500"
            >
              削除
            </button>
          </span>
        </div>

        {creatingParentId === node.id && (
          <div
            className="py-1"
            style={{ paddingLeft: `${(depth + 1) * INDENT_PX + 8}px` }}
          >
            <CollectionForm
              submitLabel="作成"
              onSubmit={async (values) => {
                await onCreate(values, node.id)
                setCreatingParentId(undefined)
              }}
              onCancel={() => setCreatingParentId(undefined)}
            />
          </div>
        )}

        {node.children.length > 0 && (
          <ul>{node.children.map((child) => renderNode(child, depth + 1))}</ul>
        )}
      </li>
    )
  }

  return (
    <aside className="w-full shrink-0 px-4 py-6 sm:w-56 sm:py-10">
      <p className="mb-8 font-display text-xl font-medium text-stone-800">
        LinkVault
      </p>

      <div className="mb-6">
        <p className="mb-1.5 px-2 text-xs font-medium tracking-wide text-stone-400">
          LIBRARY
        </p>
        <button
          type="button"
          onClick={() => onSelect(null)}
          className={`w-full rounded-lg px-2 py-1.5 text-left text-sm transition ${
            activeCollectionId === null
              ? 'bg-teal-50 font-medium text-teal-700'
              : 'text-stone-600 hover:bg-stone-100'
          }`}
        >
          すべてのリンク
        </button>
      </div>

      <div>
        <div className="mb-1.5 flex items-center justify-between px-2">
          <p className="text-xs font-medium tracking-wide text-stone-400">
            COLLECTIONS
          </p>
          <button
            type="button"
            onClick={() =>
              setCreatingParentId((v) => (v === null ? undefined : null))
            }
            className="text-stone-400 hover:text-teal-600"
            aria-label="新しいCollection"
          >
            +
          </button>
        </div>

        {creatingParentId === null && (
          <div className="mb-2 px-2">
            <CollectionForm
              submitLabel="作成"
              onSubmit={async (values) => {
                await onCreate(values, null)
                setCreatingParentId(undefined)
              }}
              onCancel={() => setCreatingParentId(undefined)}
            />
          </div>
        )}

        <ul className="flex flex-col gap-0.5">
          {tree.length === 0 && creatingParentId === undefined && (
            <li className="px-2 py-1.5 text-xs text-stone-400">
              Collectionがまだありません
            </li>
          )}
          {tree.map((node) => renderNode(node, 0))}
        </ul>
      </div>
    </aside>
  )
}

export default Sidebar
