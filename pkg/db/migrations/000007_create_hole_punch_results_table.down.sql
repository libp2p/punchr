BEGIN;

DROP INDEX IF EXISTS idx_hole_punch_results_created_at;
DROP INDEX IF EXISTS idx_hole_punch_results_x_multi_addresses_1;
DROP INDEX IF EXISTS idx_hole_punch_results_x_multi_addresses_2;
DROP TABLE IF EXISTS hole_punch_results_x_multi_addresses;
DROP TYPE IF EXISTS hole_punch_multi_address_relationship;

DROP TABLE IF EXISTS hole_punch_attempt;
DROP TYPE IF EXISTS hole_punch_attempt_outcome;

DROP TABLE IF EXISTS hole_punch_results;
DROP TYPE IF EXISTS hole_punch_outcome;

COMMIT;
