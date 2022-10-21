BEGIN;

CREATE TYPE connection_direction AS ENUM (
    'UNKNOWN',
    'OUTBOUND',
    'INBOUND'
    );

CREATE TABLE connection_events
(
    id                             INT GENERATED ALWAYS AS IDENTITY,
    local_id                       BIGINT               NOT NULL,
    remote_id                      BIGINT               NOT NULL,
    connection_multi_address_id    BIGINT               NOT NULL,
    direction                      connection_direction NOT NULL,
    listens_on_relay_multi_address BOOLEAN              NOT NULL,
    supports_dcutr                 BOOLEAN              NOT NULL,
    opened_at                      TIMESTAMPTZ          NOT NULL,
    created_at                     TIMESTAMPTZ          NOT NULL,

    CONSTRAINT fk_connection_events_local_id FOREIGN KEY (local_id) REFERENCES peers (id) ON DELETE CASCADE,
    CONSTRAINT fk_connection_events_remote_id FOREIGN KEY (remote_id) REFERENCES peers (id) ON DELETE CASCADE,
    CONSTRAINT fk_connection_events_multi_address_id FOREIGN KEY (connection_multi_address_id) REFERENCES multi_addresses (id) ON DELETE CASCADE,

    PRIMARY KEY (id)
);

CREATE INDEX idx_connection_events_opened_at ON connection_events (opened_at);

-- End the transaction
COMMIT;
