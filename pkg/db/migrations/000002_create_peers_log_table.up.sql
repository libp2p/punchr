BEGIN;
CREATE TABLE peer_logs
(
    id         INT GENERATED ALWAYS AS IDENTITY,
    peer_id    BIGINT      NOT NULL,
    field      TEXT        NOT NULL,
    old        TEXT        NOT NULL,
    new        TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,

    CONSTRAINT fk_peer_logs_peer_id FOREIGN KEY (peer_id) REFERENCES peers (id),

    PRIMARY KEY (id)
);

CREATE OR REPLACE FUNCTION insert_peer_log()
    RETURNS trigger AS
$$
BEGIN
    IF OLD.agent_version != NEW.agent_version THEN
        INSERT INTO peer_logs (peer_id, field, old, new, created_at)
        VALUES (NEW.id, 'agent_version', OLD.agent_version, NEW.agent_version, NOW());
    END IF;

    IF OLD.protocols != NEW.protocols THEN
        INSERT INTO peer_logs (peer_id, field, old, new, created_at)
        VALUES (NEW.id, 'protocols', array_to_string(OLD.protocols, ','), array_to_string(NEW.protocols, ','), NOW());
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE 'plpgsql';

CREATE TRIGGER on_peer_update
    BEFORE UPDATE
    ON peers
    FOR EACH ROW
EXECUTE PROCEDURE insert_peer_log();

END;
