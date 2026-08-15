CREATE TABLE IF NOT EXISTS collections (
    id          BIGSERIAL PRIMARY KEY,
    name        TEXT NOT NULL UNIQUE,
    description TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS collection_links (
    collection_id BIGINT NOT NULL REFERENCES collections(id) ON DELETE CASCADE,
    link_id       BIGINT NOT NULL REFERENCES links(id) ON DELETE CASCADE,
    PRIMARY KEY (collection_id, link_id)
);
