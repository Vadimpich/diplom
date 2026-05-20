CREATE TABLE system_settings (
    singleton BOOLEAN PRIMARY KEY DEFAULT TRUE CHECK (singleton),
    audio_retention_ttl_days INTEGER NOT NULL,
    processing_max_attempts INTEGER NOT NULL,
    kesmi_max_retries INTEGER NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
