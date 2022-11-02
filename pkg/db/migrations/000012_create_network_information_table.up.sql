BEGIN;

CREATE TABLE network_information
(
    id                  INT GENERATED ALWAYS AS IDENTITY,
    peer_id             BIGINT      NOT NULL,
    supports_ipv6       BOOLEAN,
    supports_ipv6_error TEXT,
    router_html         TEXT,
    router_html_error   TEXT,
    created_at          TIMESTAMPTZ NOT NULL,

    CONSTRAINT fk_network_information_peer_id FOREIGN KEY (peer_id) REFERENCES peers (id) ON DELETE CASCADE,

    PRIMARY KEY (id)
);

COMMIT;
