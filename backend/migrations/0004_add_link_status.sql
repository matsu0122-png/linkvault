ALTER TABLE links
    ADD COLUMN IF NOT EXISTS status TEXT NOT NULL DEFAULT 'unknown' CHECK (status IN ('unknown', 'ok', 'broken')),
    ADD COLUMN IF NOT EXISTS checked_at TIMESTAMPTZ;
