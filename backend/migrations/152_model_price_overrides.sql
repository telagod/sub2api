-- 152_model_price_overrides.sql
-- 管理员对模型「官方价基准」的覆盖（OpenRouter 定价取舍）。
-- 优先级：手动价(manual_*) > 指定供应商(pinned_provider_tag) > 自动最低价。
-- model_id 为匹配键（本地模型名或 OpenRouter slug）；硬删除即恢复自动。

CREATE TABLE IF NOT EXISTS model_price_overrides (
    id                    BIGSERIAL PRIMARY KEY,
    model_id              VARCHAR(200) NOT NULL,
    pinned_provider_tag   VARCHAR(120) NOT NULL DEFAULT '',
    manual_input          DOUBLE PRECISION,
    manual_output         DOUBLE PRECISION,
    manual_cache_read     DOUBLE PRECISION,
    manual_cache_write    DOUBLE PRECISION,
    note                  VARCHAR(500) NOT NULL DEFAULT '',
    updated_by            BIGINT NOT NULL DEFAULT 0,
    created_at            TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at            TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS modelpriceoverride_model_id_unique
    ON model_price_overrides (model_id);
