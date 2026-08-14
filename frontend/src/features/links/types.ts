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
