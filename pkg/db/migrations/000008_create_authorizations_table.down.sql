BEGIN;

DROP INDEX IF EXISTS idx_authorizations_api_key;
DROP TABLE IF EXISTS authorizations;

COMMIT;
