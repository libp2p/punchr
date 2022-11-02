BEGIN;

CREATE TYPE latency_measurement_type AS ENUM (
    'TO_RELAY',
    'TO_REMOTE_THROUGH_RELAY',
    'TO_REMOTE_AFTER_HOLEPUNCH'
    );

CREATE TABLE latency_measurements
(
    id                   INT GENERATED ALWAYS AS IDENTITY,
    remote_id            BIGINT                   NOT NULL,

    hole_punch_result_id INT                      NOT NULL,
    multi_address_id     BIGINT                   NOT NULL,
    mtype                latency_measurement_type NOT NULL,

    rtts                 float[]                  NOT NULL,
    rtt_avg              float                    NOT NULL,
    rtt_max              float                    NOT NULL,
    rtt_min              float                    NOT NULL,
    rtt_std              float                    NOT NULL,

    CONSTRAINT fk_latency_measurements_remote_id FOREIGN KEY (remote_id) REFERENCES peers (id) ON DELETE CASCADE,
    CONSTRAINT fk_latency_measurements_multi_address_id FOREIGN KEY (multi_address_id) REFERENCES multi_addresses (id) ON DELETE CASCADE,
    CONSTRAINT fk_latency_measurements_hole_punch_result_id FOREIGN KEY (hole_punch_result_id) REFERENCES hole_punch_results (id) ON DELETE CASCADE,

    PRIMARY KEY (id)
);

COMMIT;
