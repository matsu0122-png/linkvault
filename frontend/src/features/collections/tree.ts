import type { Collection } from './types'

export type CollectionTreeNode = Collection & {
  children: CollectionTreeNode[]
}

// parent_idが指す先が一覧に存在しない場合(例: 親削除直後の一時的な
// ローカル状態のずれ)は、そのCollectionを最上位として扱う。
export function buildCollectionTree(
  collections: Collection[],
): CollectionTreeNode[] {
  const nodes = new Map<number, CollectionTreeNode>()
  collections.forEach((c) => nodes.set(c.id, { ...c, children: [] }))

  const roots: CollectionTreeNode[] = []
  nodes.forEach((node) => {
    const parent =
      node.parent_id !== null ? nodes.get(node.parent_id) : undefined
    if (parent) {
      parent.children.push(node)
    } else {
      roots.push(node)
    }
  })

  function sortTree(list: CollectionTreeNode[]) {
    list.sort((a, b) => a.name.localeCompare(b.name))
    list.forEach((n) => sortTree(n.children))
  }
  sortTree(roots)

  return roots
}

export type FlatCollectionNode = { collection: Collection; depth: number }

// ツリーを「親の直後に子」という深さ優先の並びへ平坦化する。
// チェックリストのようなインデント付きフラットリストUIで使う。
export function flattenCollectionTree(
  tree: CollectionTreeNode[],
  depth = 0,
): FlatCollectionNode[] {
  return tree.flatMap((node) => [
    { collection: node, depth },
    ...flattenCollectionTree(node.children, depth + 1),
  ])
}
