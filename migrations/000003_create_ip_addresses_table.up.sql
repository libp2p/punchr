CREATE TABLE ip_addresses
(
    id         BIGINT GENERATED ALWAYS AS IDENTITY,
    address    INET        NOT NULL,
    country    VARCHAR(2),
    continent  VARCHAR(2),
    asn        INT,
    is_public  BOOL        NOT NULL,

    updated_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,

    CONSTRAINT uq_ip_addresses_address UNIQUE (address),

    PRIMARY KEY (id)
);

