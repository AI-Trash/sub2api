-- Compatibility migration for the legacy 108 webhook migration.
-- The original 108_add_ops_alert_webhook_notifications.sql used goose Up/Down
-- markers, but this repository's runner executes the whole file as-is. Keep
-- this forward-only migration so environments that recorded 108 still receive
-- the intended schema change.
ALTER TABLE ops_alert_rules
    ADD COLUMN IF NOT EXISTS notify_webhook BOOLEAN NOT NULL DEFAULT false;

ALTER TABLE ops_alert_events
    ADD COLUMN IF NOT EXISTS webhook_sent BOOLEAN NOT NULL DEFAULT false;
