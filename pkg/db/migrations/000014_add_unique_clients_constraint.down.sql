BEGIN;

ALTER TABLE clients DROP CONSTRAINT uq_clients_id_peer_id;

COMMIT;
