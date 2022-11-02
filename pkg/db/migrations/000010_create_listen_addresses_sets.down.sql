BEGIN;

ALTER TABLE hole_punch_results
    DROP COLUMN listen_multi_addresses_set_id;
DROP FUNCTION IF EXISTS upsert_multi_addresses_sets;
DROP TABLE IF EXISTS multi_addresses_sets;
DROP EXTENSION IF EXISTS intarray;

COMMIT;
