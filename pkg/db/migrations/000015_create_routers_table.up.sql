BEGIN;

CREATE TABLE routers
(
    id         INT GENERATED ALWAYS AS IDENTITY,
    client_id  INT         NOT NULL,
    html       TEXT        NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,

    CONSTRAINT fk_routers_client_id FOREIGN KEY (client_id) REFERENCES peers (id) ON DELETE CASCADE,

    PRIMARY KEY (id)
);

COMMIT;
