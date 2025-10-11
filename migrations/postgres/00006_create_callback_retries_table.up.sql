-- Create callback_retries table for webhook retry mechanism
CREATE TABLE IF NOT EXISTS callback_retries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    payment_id UUID NOT NULL,
    callback_data JSONB NOT NULL,
    retry_count INTEGER NOT NULL DEFAULT 0,
    max_retries INTEGER NOT NULL DEFAULT 5,
    last_error TEXT,
    next_retry_at TIMESTAMPTZ,
    status VARCHAR(50) NOT NULL DEFAULT 'pending', -- pending, processing, completed, failed
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_payment FOREIGN KEY (payment_id) REFERENCES payments(id) ON DELETE CASCADE
);

-- Create index on payment_id for faster lookups
CREATE INDEX idx_callback_retries_payment_id ON callback_retries(payment_id);

-- Create index on status and next_retry_at for efficient retry queue processing
CREATE INDEX idx_callback_retries_status_next_retry ON callback_retries(status, next_retry_at) WHERE status = 'pending';

-- Create index on created_at for cleanup queries
CREATE INDEX idx_callback_retries_created_at ON callback_retries(created_at);

-- Add trigger to auto-update updated_at
CREATE OR REPLACE FUNCTION update_callback_retries_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_callback_retries_updated_at
    BEFORE UPDATE ON callback_retries
    FOR EACH ROW
    EXECUTE FUNCTION update_callback_retries_updated_at();
