export type Collection = {
  id: number
  name: string
  description: string
  parent_id: number | null
  link_count: number
  created_at: string
  updated_at: string
}

export type CreateCollectionInput = {
  name: string
  description: string
  parent_id: number | null
}

// parent_idはPUTでは変更できない（バックエンドが無視する）ため含めない。
export type UpdateCollectionInput = {
  name: string
  description: string
}
