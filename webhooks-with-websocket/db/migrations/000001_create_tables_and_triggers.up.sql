CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE subscriptions
(
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4()
);

CREATE TABLE webhooks
(
    id              uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    subscription_id uuid  NOT NULL,
    -- created_at must be text, since the row_to_json drops the timestamp --
    created_at      text  NOT NULL   DEFAULT CURRENT_TIMESTAMP,
    acked_at        TIMESTAMP,
    payload         JSONB NOT NULL,
    CONSTRAINT fk_subscription FOREIGN KEY (subscription_id) REFERENCES subscriptions (id) ON DELETE CASCADE
);


-- CREATE FUNCTION --


CREATE OR REPLACE FUNCTION notify_webhook() RETURNS TRIGGER AS
$$

DECLARE
    notification json;

BEGIN
    IF (TG_OP = 'INSERT') THEN
        -- Contruct the notification as a JSON string.
        notification = row_to_json(NEW);

        -- Execute pg_notify(channel, notification)
        PERFORM pg_notify('webhook_created', notification::text);
    END IF;

    RETURN NULL;
END;

$$ LANGUAGE plpgsql;

-- CREATE TRIGGER --

CREATE TRIGGER notify_webhook
    AFTER INSERT OR UPDATE OR DELETE
    ON webhooks
    FOR EACH ROW
EXECUTE PROCEDURE notify_webhook();

INSERT INTO subscriptions (id)
VALUES ('a2cce679-0b59-4245-a389-298a423945c0');
