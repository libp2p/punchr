BEGIN;

CREATE TABLE connection_events_x_multi_addresses
(
    connection_event_id INT    NOT NULL,
    multi_address_id    BIGINT NOT NULL,

    CONSTRAINT fk_connection_events_x_multi_addresses_multi_address_id FOREIGN KEY (multi_address_id) REFERENCES multi_addresses (id) ON DELETE CASCADE,
    CONSTRAINT fk_connection_events_x_multi_addresses_connection_event_id FOREIGN KEY (connection_event_id) REFERENCES connection_events (id) ON DELETE CASCADE,

    PRIMARY KEY (multi_address_id, connection_event_id)
);

CREATE INDEX idx_connection_events_x_multi_addresses_1 ON connection_events_x_multi_addresses (connection_event_id, multi_address_id);
CREATE INDEX idx_connection_events_x_multi_addresses_2 ON connection_events_x_multi_addresses (multi_address_id, connection_event_id);

COMMIT;
