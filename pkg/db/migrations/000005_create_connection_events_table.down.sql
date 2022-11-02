BEGIN;
DROP INDEX idx_connection_events_opened_at;
DROP TABLE connection_events;
COMMIT;
