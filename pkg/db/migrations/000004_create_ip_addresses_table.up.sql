-- Rows in the `ip_addresses` capture information for a particular IP address
-- that were derived from a multi address
CREATE TABLE ip_addresses
(
    -- A unique id that identifies this ip address.
    id               INT GENERATED ALWAYS AS IDENTITY,
    -- The multi address that this ip address belongs to.
    multi_address_id BIGINT      NOT NULL,
    -- The autonomous system number that this ip address belongs to.
    asn              INT,
    -- If NULL this address could not be associated with a cloud provider.
    -- If not NULL the integer corresponds to the UdgerDB datacenter ID.
    is_cloud         INT,
    -- When was this IP address updated
    updated_at       TIMESTAMPTZ NOT NULL,
    -- When was this IP address created
    created_at       TIMESTAMPTZ NOT NULL,
    -- The country that this address belongs to in the form of a two to three letter country code
    country          CHAR(2) CHECK ( TRIM(country) != '' ), -- make it not null so that the unique constraint applies IPs without country.
    -- The continent that this address belongs to in the form of a two letter code.
    continent        CHAR(2) CHECK ( TRIM(continent) != '' ),
    -- The IP address derived from the reference multi address.
    address          INET        NOT NULL,


    -- Only one address/multi_address_id combination should be allowed.
    CONSTRAINT uq_ip_addresses_multi_address_id_address UNIQUE (multi_address_id, address),

    -- The multi_address_id should reference the proper table row.
    CONSTRAINT fk_ip_addresses_multi_address_id FOREIGN KEY (multi_address_id) REFERENCES multi_addresses (id) ON DELETE CASCADE,

    PRIMARY KEY (id)
);
