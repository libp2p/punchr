BEGIN;

CREATE TYPE connection_direction AS ENUM (
    'UNKNOWN',
    'OUTBOUND',
    'INBOUND'
    );

CREATE TABLE connection_events
(
    id               INT GENERATED ALWAYS AS IDENTITY,
    local_id         BIGINT               NOT NULL,
    remote_id        BIGINT               NOT NULL,
    multi_address_id BIGINT               NOT NULL,
    direction        connection_direction NOT NULL,
    relayed          BOOLEAN              NOT NULL,
    opened_at        TIMESTAMPTZ          NOT NULL,
    created_at       TIMESTAMPTZ          NOT NULL,

    CONSTRAINT fk_connection_events_local_id FOREIGN KEY (local_id) REFERENCES peers (id) ON DELETE CASCADE,
    CONSTRAINT fk_connection_events_remote_id FOREIGN KEY (remote_id) REFERENCES peers (id) ON DELETE CASCADE,
    CONSTRAINT fk_connection_events_multi_address_id FOREIGN KEY (multi_address_id) REFERENCES multi_addresses (id) ON DELETE CASCADE,

    PRIMARY KEY (id)
);

-- End the transaction
COMMIT;
