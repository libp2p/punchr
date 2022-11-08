BEGIN;

CREATE TABLE port_mappings
(
    id                   INT GENERATED ALWAYS AS IDENTITY,
    hole_punch_result_id INT  NOT NULL,
    internal_port        INT  NOT NULL,
    external_port        INT  NOT NULL,
    protocol             TEXT NOT NULL,
    addr                 TEXT NOT NULL,
    addr_network         TEXT NOT NULL,

    CONSTRAINT fk_latency_measurements_hole_punch_result_id FOREIGN KEY (hole_punch_result_id) REFERENCES hole_punch_results (id) ON DELETE CASCADE,

    PRIMARY KEY (id)
);

COMMIT;
