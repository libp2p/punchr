BEGIN;

CREATE TYPE hole_punch_outcome AS ENUM (
    'UNKNOWN',
    'NO_CONNECTION',
    'NO_STREAM',
    'CONNECTION_REVERSED',
    'CANCELLED',
    'FAILED',
    'SUCCESS'
    );

CREATE TABLE hole_punch_results
(
    id                 INT GENERATED ALWAYS AS IDENTITY,
    client_id          BIGINT             NOT NULL,
    remote_id          BIGINT             NOT NULL,
    connect_started_at TIMESTAMPTZ        NOT NULL,
    connect_ended_at   TIMESTAMPTZ        NOT NULL,
    has_direct_conns   BOOLEAN            NOT NULL,
    error              TEXT,
    outcome            hole_punch_outcome NOT NULL,
    ended_at           TIMESTAMPTZ        NOT NULL,
    updated_at         TIMESTAMPTZ        NOT NULL,
    created_at         TIMESTAMPTZ        NOT NULL,

    CONSTRAINT fk_hole_punch_results_client_id FOREIGN KEY (client_id) REFERENCES peers (id) ON DELETE CASCADE,
    CONSTRAINT fk_hole_punch_results_remote_id FOREIGN KEY (remote_id) REFERENCES peers (id) ON DELETE CASCADE,

    PRIMARY KEY (id)
);

CREATE TYPE hole_punch_attempt_outcome AS ENUM (
    'UNKNOWN',
    'DIRECT_DIAL',
    'PROTOCOL_ERROR',
    'CANCELLED',
    'TIMEOUT',
    'FAILED',
    'SUCCESS'
    );

CREATE TABLE hole_punch_attempt
(
    id                   INT GENERATED ALWAYS AS IDENTITY,
    hole_punch_result_id INT                        NOT NULL,
    opened_at            TIMESTAMPTZ                NOT NULL,
    started_at           TIMESTAMPTZ,
    ended_at             TIMESTAMPTZ                NOT NULL,
    start_rtt            INTERVAL,
    elapsed_time         INTERVAL                   NOT NULL,
    outcome              hole_punch_attempt_outcome NOT NULL,
    error                TEXT,
    direct_dial_error    TEXT,

    updated_at           TIMESTAMPTZ                NOT NULL,
    created_at           TIMESTAMPTZ                NOT NULL,

    CONSTRAINT fk_hole_punch_attempt_hole_punch_result_id FOREIGN KEY (hole_punch_result_id) REFERENCES hole_punch_results (id) ON DELETE CASCADE,

    PRIMARY KEY (id)
);


CREATE TYPE hole_punch_multi_address_relationship AS ENUM (
    -- INITIAL multi addresses are the multi addresses that correspond
    -- to the open connections to the remote peer BEFORE the hole punch
    -- (the ones that were used to connect to the remote peer, usually
    -- only one relay address)
    'INITIAL',
    -- FINAL multi addresses are the multi addresses that correspond
    -- to the open connections to the remote peer AFTER the hole punch.
    -- Usually this includes the relay + a direct connection to the
    -- remote peer.
    'FINAL'
    );

CREATE TABLE hole_punch_results_x_multi_addresses
(
    hole_punch_result_id INT                                   NOT NULL,
    multi_address_id     BIGINT                                NOT NULL,
    relationship         hole_punch_multi_address_relationship NOT NULL,

    CONSTRAINT fk_hole_punch_results_x_multi_addresses_maddr_id FOREIGN KEY (multi_address_id) REFERENCES multi_addresses (id) ON DELETE CASCADE,
    CONSTRAINT fk_hole_punch_results_x_multi_addresses_hpr_id FOREIGN KEY (hole_punch_result_id) REFERENCES hole_punch_results (id) ON DELETE CASCADE,

    PRIMARY KEY (multi_address_id, hole_punch_result_id, relationship)
);

CREATE INDEX idx_hole_punch_results_x_multi_addresses_1 ON hole_punch_results_x_multi_addresses (hole_punch_result_id, multi_address_id);
CREATE INDEX idx_hole_punch_results_x_multi_addresses_2 ON hole_punch_results_x_multi_addresses (multi_address_id, hole_punch_result_id);

-- End the transaction
COMMIT;
