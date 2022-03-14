BEGIN;

CREATE TYPE hole_punch_end_reason AS ENUM (
    'direct_dial',
    'protocol_error',
    'hole_punch'
    );

CREATE TABLE hole_punch_results
(
    id                INT GENERATED ALWAYS AS IDENTITY,
    local_id          BIGINT                NOT NULL,
    remote_id         BIGINT                NOT NULL,
    multi_address_id  BIGINT                NOT NULL,
    start_rtt         INTERVAL              NOT NULL,
    elapsed_time      INTERVAL              NOT NULL,
    end_reason        hole_punch_end_reason NOT NULL,
    attempts          SMALLINT              NOT NULL,
    success           BOOLEAN               NOT NULL,
    error             TEXT,
    direct_dial_error TEXT,
    updated_at        TIMESTAMPTZ           NOT NULL,
    created_at        TIMESTAMPTZ           NOT NULL,

    CONSTRAINT fk_connection_events_local_id FOREIGN KEY (local_id) REFERENCES peers (id) ON DELETE CASCADE,
    CONSTRAINT fk_connection_events_remote_id FOREIGN KEY (remote_id) REFERENCES peers (id) ON DELETE CASCADE,
    CONSTRAINT fk_connection_events_multi_address_id FOREIGN KEY (multi_address_id) REFERENCES multi_addresses (id) ON DELETE CASCADE,

    PRIMARY KEY (id)
);

-- End the transaction
COMMIT;
