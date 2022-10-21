CREATE TABLE multi_addresses
(
    id               BIGINT GENERATED ALWAYS AS IDENTITY,
    maddr            TEXT        NOT NULL,
    country          VARCHAR(2),
    continent        VARCHAR(2),
    asn              INT,
    is_public        BOOL        NOT NULL,
    is_relay         BOOL        NOT NULL,
    ip_address_count INT         NOT NULL,

    updated_at       TIMESTAMPTZ NOT NULL,
    created_at       TIMESTAMPTZ NOT NULL,

    CONSTRAINT uq_multi_addresses_maddr UNIQUE (maddr),

    PRIMARY KEY (id)
);
