BEGIN;

CREATE TABLE clients
(

    id               INT GENERATED ALWAYS AS IDENTITY,
    peer_id          BIGINT NOT NULL,
    authorization_id INT    NOT NULL,

    CONSTRAINT fk_clients_client_id FOREIGN KEY (peer_id) REFERENCES peers (id) ON DELETE CASCADE,
    CONSTRAINT fk_clients_authorization_id FOREIGN KEY (authorization_id) REFERENCES authorizations (id) ON DELETE CASCADE,

    PRIMARY KEY (id)
);

CREATE INDEX idx_clients_peer_id ON clients (peer_id);

COMMIT;
