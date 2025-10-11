CREATE TABLE IF NOT EXISTS payments (
    id                      UUID PRIMARY KEY DEFAULT GEN_RANDOM_UUID(),
    invoice_id              VARCHAR(100) NOT NULL UNIQUE,
    member_id               UUID NOT NULL REFERENCES members (id) ON DELETE CASCADE,
    amount                  BIGINT NOT NULL CHECK (amount > 0),
    currency                VARCHAR(3) NOT NULL,
    status                  VARCHAR(20) NOT NULL DEFAULT 'pending',
    payment_method          VARCHAR(20) NOT NULL DEFAULT 'card',
    payment_type            VARCHAR(20) NOT NULL,
    related_entity_id       UUID,
    gateway_transaction_id  VARCHAR(100),
    gateway_response        TEXT,
    card_mask               VARCHAR(20),
    approval_code           VARCHAR(50),
    error_code              VARCHAR(50),
    error_message           TEXT,
    created_at              TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at              TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    completed_at            TIMESTAMP,
    expires_at              TIMESTAMP NOT NULL,

    CONSTRAINT status_check CHECK (status IN ('pending', 'processing', 'completed', 'failed', 'cancelled', 'refunded')),
    CONSTRAINT payment_method_check CHECK (payment_method IN ('card', 'wallet')),
    CONSTRAINT payment_type_check CHECK (payment_type IN ('fine', 'subscription', 'deposit'))
);

-- Index for faster queries by member
CREATE INDEX idx_payments_member_id ON payments(member_id);

-- Index for faster queries by invoice ID
CREATE INDEX idx_payments_invoice_id ON payments(invoice_id);

-- Index for faster queries by status
CREATE INDEX idx_payments_status ON payments(status);

-- Index for faster queries by gateway transaction ID
CREATE INDEX idx_payments_gateway_transaction_id ON payments(gateway_transaction_id) WHERE gateway_transaction_id IS NOT NULL;

-- Index for faster queries by payment type
CREATE INDEX idx_payments_type ON payments(payment_type);

-- Index for finding pending/processing payments
CREATE INDEX idx_payments_active ON payments(status, expires_at) WHERE status IN ('pending', 'processing');

-- Index for finding completed payments
CREATE INDEX idx_payments_completed ON payments(completed_at) WHERE status = 'completed';

-- Add trigger for updating updated_at timestamp
CREATE OR REPLACE FUNCTION update_payment_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER payments_updated_at
    BEFORE UPDATE ON payments
    FOR EACH ROW
    EXECUTE FUNCTION update_payment_updated_at();
