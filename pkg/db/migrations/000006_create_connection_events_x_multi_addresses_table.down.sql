BEGIN;

DROP INDEX idx_connection_events_x_multi_addresses_1;
DROP INDEX idx_connection_events_x_multi_addresses_2;
DROP TABLE connection_events_x_multi_addresses;

COMMIT;
