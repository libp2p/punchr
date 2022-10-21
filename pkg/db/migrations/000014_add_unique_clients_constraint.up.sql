BEGIN;

ALTER TABLE clients ADD CONSTRAINT uq_clients_id_peer_id UNIQUE (id, peer_id);

COMMIT;
