BEGIN;

CREATE TABLE multi_addresses_x_ip_addresses
(
    multi_address_id BIGINT NOT NULL,
    ip_address_id    BIGINT NOT NULL,

    CONSTRAINT fk_multi_addresses_x_ip_addresses_multi_address_ip FOREIGN KEY (multi_address_id) REFERENCES multi_addresses (id) ON DELETE CASCADE,
    CONSTRAINT fk_multi_addresses_x_ip_addresses_ip_address_id FOREIGN KEY (ip_address_id) REFERENCES ip_addresses (id) ON DELETE CASCADE,

    PRIMARY KEY (multi_address_id, ip_address_id)
);

CREATE INDEX idx_multi_addresses_x_ip_addresses_1 ON multi_addresses_x_ip_addresses (ip_address_id, multi_address_id);
CREATE INDEX idx_multi_addresses_x_ip_addresses_2 ON multi_addresses_x_ip_addresses (multi_address_id, ip_address_id);

-- End the transaction
COMMIT;
