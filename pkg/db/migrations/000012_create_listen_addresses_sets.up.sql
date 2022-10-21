BEGIN;

CREATE EXTENSION intarray;

CREATE TABLE multi_addresses_sets
(

    id                  INT GENERATED ALWAYS AS IDENTITY,
    multi_addresses_ids INT[] NOT NULL,
    digest              BYTEA NOT NULL,

    CONSTRAINT uq_multi_addresses_sets_digest UNIQUE (digest),

    PRIMARY KEY (id)
);

CREATE OR REPLACE FUNCTION upsert_multi_addresses_sets(
    new_multi_addresses_ids INT[]
) RETURNS INT AS
$upsert_multi_addresses_sets$
DECLARE
    sorted_new_multi_addresses INT[];
    new_multi_addresses_digest BYTEA;
    multi_addresses_set_id     INT;
    multi_addresses_set        multi_addresses_sets%rowtype;
BEGIN
    IF new_multi_addresses_ids IS NULL OR array_length(new_multi_addresses_ids, 1) IS NULL THEN
        RETURN NULL;
    END IF;

    SELECT uniq(sort_asc(new_multi_addresses_ids)) INTO sorted_new_multi_addresses;
    SELECT sha256(array_to_string(sorted_new_multi_addresses, ':')::BYTEA) INTO new_multi_addresses_digest;

    SELECT *
    FROM multi_addresses_sets mas
    WHERE mas.digest = new_multi_addresses_digest
    INTO multi_addresses_set;

    IF multi_addresses_set IS NULL THEN
        INSERT INTO multi_addresses_sets (multi_addresses_ids, digest)
        VALUES (sorted_new_multi_addresses, new_multi_addresses_digest)
        RETURNING id INTO multi_addresses_set_id;

        RETURN multi_addresses_set_id;
    END IF;

    RETURN multi_addresses_set.id;
END;
$upsert_multi_addresses_sets$ LANGUAGE plpgsql;

ALTER TABLE hole_punch_results
    ADD COLUMN listen_multi_addresses_set_id INT NOT NULL DEFAULT 0;
ALTER TABLE hole_punch_results
    ADD CONSTRAINT fk_hole_punch_results_listen_multi_addresses_set_id FOREIGN KEY (listen_multi_addresses_set_id) REFERENCES multi_addresses_sets (id);

COMMIT;
