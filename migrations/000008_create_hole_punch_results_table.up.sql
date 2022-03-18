BEGIN;

CREATE TYPE hole_punch_end_reason AS ENUM (
    'UNKNOWN',
    'NO_CONNECTION',
    'DIRECT_CONNECTION',
    'NOT_INITIATED',
    'DIRECT_DIAL',
    'PROTOCOL_ERROR',
    'HOLE_PUNCH'
    );

CREATE TABLE hole_punch_results
(
    id                    INT GENERATED ALWAYS AS IDENTITY,
    client_id             BIGINT                NOT NULL,
    remote_id             BIGINT                NOT NULL,
    connection_started_at TIMESTAMPTZ           NOT NULL,
    start_rtt             INTERVAL,
    elapsed_time          INTERVAL              NOT NULL,
    end_reason            hole_punch_end_reason NOT NULL,
    attempts              SMALLINT              NOT NULL,
    success               BOOLEAN               NOT NULL,
    error                 TEXT,
    direct_dial_error     TEXT,
    updated_at            TIMESTAMPTZ           NOT NULL,
    created_at            TIMESTAMPTZ           NOT NULL,

    CONSTRAINT fk_connection_events_client_id FOREIGN KEY (client_id) REFERENCES peers (id) ON DELETE CASCADE,
    CONSTRAINT fk_connection_events_remote_id FOREIGN KEY (remote_id) REFERENCES peers (id) ON DELETE CASCADE,

    PRIMARY KEY (id)
);

-- End the transaction
COMMIT;
