-- Allow channel model pricing to override per-service-tier billing multipliers.

ALTER TABLE channel_model_pricing
    ADD COLUMN IF NOT EXISTS service_tier_multipliers JSONB NOT NULL DEFAULT '{}'::jsonb;

COMMENT ON COLUMN channel_model_pricing.service_tier_multipliers IS '模型 service tier 费用倍率覆盖，JSON 对象：{"priority": 2, "flex": 0.5}；未配置时沿用默认计费逻辑';
