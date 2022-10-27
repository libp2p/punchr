BEGIN;

CREATE TABLE connection_events
(
    -- A unique ID of this connection event
    id               INT GENERATED ALWAYS AS IDENTITY,
    -- The local peer ID of the honeypot
    local_id         BIGINT      NOT NULL,
    -- The peer ID of the remote peer
    remote_id        BIGINT      NOT NULL,
    -- The multi address of the connection
    multi_address_id BIGINT      NOT NULL,
    -- When was this connection opened
    opened_at        TIMESTAMPTZ NOT NULL,
    -- When was this event written to the DB
    created_at       TIMESTAMPTZ NOT NULL,

    CONSTRAINT fk_connection_events_local_id FOREIGN KEY (local_id) REFERENCES peers (id) ON DELETE CASCADE,
    CONSTRAINT fk_connection_events_remote_id FOREIGN KEY (remote_id) REFERENCES peers (id) ON DELETE CASCADE,
    CONSTRAINT fk_connection_events_multi_address_id FOREIGN KEY (multi_address_id) REFERENCES multi_addresses (id) ON DELETE CASCADE,

    PRIMARY KEY (id)
);

CREATE INDEX idx_connection_events_opened_at ON connection_events (opened_at);

-- End the transaction
COMMIT;
