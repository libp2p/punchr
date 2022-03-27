BEGIN;

DROP INDEX idx_hole_punch_results_x_multi_addresses_1;
DROP INDEX idx_hole_punch_results_x_multi_addresses_2;
DROP TABLE hole_punch_results_x_multi_addresses;
DROP TYPE hole_punch_multi_address_relationship;

DROP TABLE hole_punch_attempt;
DROP TYPE hole_punch_attempt_outcome;

DROP TABLE hole_punch_results;
DROP TYPE hole_punch_outcome;


COMMIT;
