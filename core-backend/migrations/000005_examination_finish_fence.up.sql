CREATE TABLE examination_processing_launches (
    examination_id BIGINT PRIMARY KEY REFERENCES examinations (id) ON DELETE CASCADE,
    launched_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
