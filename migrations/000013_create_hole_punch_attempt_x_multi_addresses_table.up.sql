BEGIN;

CREATE TABLE hole_punch_attempt_x_multi_addresses
(
    hole_punch_attempt INT    NOT NULL,
    multi_address_id   BIGINT NOT NULL,

    CONSTRAINT fk_hole_punch_attempt_x_multi_addresses_multi_address_id FOREIGN KEY (multi_address_id) REFERENCES multi_addresses (id) ON DELETE CASCADE,
    CONSTRAINT fk_hole_punch_attempt_x_multi_addresses_hole_punch_attempt FOREIGN KEY (hole_punch_attempt) REFERENCES hole_punch_attempt (id) ON DELETE CASCADE,

    PRIMARY KEY (multi_address_id, hole_punch_attempt)
);

CREATE INDEX idx_hole_punch_attempt_x_multi_addresses ON hole_punch_attempt_x_multi_addresses (hole_punch_attempt, multi_address_id);

COMMIT;
