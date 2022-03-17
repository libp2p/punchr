BEGIN;

CREATE TABLE hole_punch_results_x_multi_addresses
(
    hole_punch_result_id INT    NOT NULL,
    multi_address_id     BIGINT NOT NULL,

    CONSTRAINT fk_hole_punch_results_x_multi_addresses_multi_address_id FOREIGN KEY (multi_address_id) REFERENCES multi_addresses (id) ON DELETE CASCADE,
    CONSTRAINT fk_hole_punch_results_x_multi_addresses_hole_punch_result_id FOREIGN KEY (hole_punch_result_id) REFERENCES hole_punch_results (id) ON DELETE CASCADE,

    PRIMARY KEY (multi_address_id, hole_punch_result_id)
);

CREATE INDEX idx_hole_punch_results_x_multi_addresses_1 ON hole_punch_results_x_multi_addresses (hole_punch_result_id, multi_address_id);
CREATE INDEX idx_hole_punch_results_x_multi_addresses_2 ON hole_punch_results_x_multi_addresses (multi_address_id, hole_punch_result_id);

COMMIT;
