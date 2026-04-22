-- +goose Up
-- +goose StatementBegin
ALTER TABLE ops_alert_rules
    ADD COLUMN IF NOT EXISTS notify_webhook BOOLEAN NOT NULL DEFAULT false;

ALTER TABLE ops_alert_events
    ADD COLUMN IF NOT EXISTS webhook_sent BOOLEAN NOT NULL DEFAULT false;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE ops_alert_events
    DROP COLUMN IF EXISTS webhook_sent;

ALTER TABLE ops_alert_rules
    DROP COLUMN IF EXISTS notify_webhook;
-- +goose StatementEnd
