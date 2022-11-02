CREATE TABLE multi_addresses
(
    -- An internal unique id that identifies this multi address.
    id             BIGINT GENERATED ALWAYS AS IDENTITY,
    -- The autonomous system number that this multi address belongs to.
    asn            INT,
    -- If NULL this multi address could not be associated with a cloud provider.
    -- If not NULL the integer corresponds to the UdgerDB datacenter ID.
    is_cloud       INT,
    -- A boolean value that indicates whether this multi address is a relay address.
    is_relay       BOOLEAN,
    -- A boolean value that indicates whether this multi address is a publicly reachable one.
    is_public      BOOLEAN,
    -- The derived IPv4 or IPv6 address that was used to determine the country etc.
    addr           INET,
    -- Indicates if the multi_address has multiple IP addresses. Could happen for dnsaddr multi addresses.
    -- If this flag is true there are corresponding IP addresses.
    has_many_addrs BOOLEAN,
    -- The country that this multi address belongs to in the form of a two letter country code.
    country        CHAR(2) CHECK ( TRIM(country) != '' ),
    -- The continent that this multi address belongs to in the form of a two letter code.
    continent      CHAR(2) CHECK ( TRIM(continent) != '' ),
    -- The multi address in the form of `/ip4/123.456.789.123/tcp/4001`.
    maddr          TEXT        NOT NULL CHECK ( TRIM(maddr) != '' ),

    -- When was this multi address updated the last time
    updated_at     TIMESTAMPTZ NOT NULL CHECK ( updated_at >= created_at ),
    -- When was this multi address created
    created_at     TIMESTAMPTZ NOT NULL,

    -- There should only ever be distinct multi addresses here
    CONSTRAINT uq_multi_addresses_address UNIQUE (maddr),

    PRIMARY KEY (id)
);
