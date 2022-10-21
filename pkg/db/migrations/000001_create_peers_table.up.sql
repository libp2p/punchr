-- The `peers` table keeps track of all peers ever seen
CREATE TABLE peers
(
    id             BIGINT GENERATED ALWAYS AS IDENTITY,
    multi_hash     TEXT        NOT NULL,
    agent_version  TEXT,
    protocols      TEXT[],
    supports_dcutr BOOLEAN     NOT NULL,
    updated_at     TIMESTAMPTZ NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL,

    CONSTRAINT uq_peers_multi_hash UNIQUE (multi_hash),

    PRIMARY KEY (id)
);
