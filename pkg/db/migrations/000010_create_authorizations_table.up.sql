BEGIN;

CREATE TABLE authorizations
(
    id         INT GENERATED ALWAYS AS IDENTITY,
    api_key    VARCHAR(36) NOT NULL,
    username   TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,

    PRIMARY KEY (id)
);

CREATE INDEX idx_authorizations_api_key ON authorizations (api_key);

COMMIT;
