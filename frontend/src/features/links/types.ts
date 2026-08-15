export type LinkStatus = 'unknown' | 'ok' | 'broken'

export type Link = {
  id: number
  url: string
  title: string
  memo: string
  tags: string[]
  description: string
  image_url: string
  favicon_url: string
  status: LinkStatus
  checked_at: string | null
  created_at: string
  updated_at: string
}

export type SortOption =
  | 'created_at_desc'
  | 'created_at_asc'
  | 'updated_at_desc'
  | 'updated_at_asc'
  | 'title_asc'
  | 'title_desc'

export type Pagination = {
  page: number
  limit: number
  total: number
  totalPages: number
}

export type LinksResponse = {
  links: Link[]
  pagination: Pagination
}
