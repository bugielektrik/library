CREATE TABLE IF NOT EXISTS saved_cards (
    id                  UUID PRIMARY KEY DEFAULT GEN_RANDOM_UUID(),
    member_id           UUID NOT NULL REFERENCES members (id) ON DELETE CASCADE,
    card_token          VARCHAR(100) NOT NULL UNIQUE,
    card_mask           VARCHAR(20) NOT NULL,
    card_type           VARCHAR(20) NOT NULL,
    expiry_month        INT NOT NULL CHECK (expiry_month >= 1 AND expiry_month <= 12),
    expiry_year         INT NOT NULL,
    is_default          BOOLEAN NOT NULL DEFAULT FALSE,
    is_active           BOOLEAN NOT NULL DEFAULT TRUE,
    created_at          TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at          TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_used_at        TIMESTAMP,

    CONSTRAINT valid_expiry_month CHECK (expiry_month BETWEEN 1 AND 12)
);

-- Index for faster queries by member
CREATE INDEX idx_saved_cards_member_id ON saved_cards(member_id);

-- Index for finding default cards
CREATE INDEX idx_saved_cards_default ON saved_cards(member_id, is_default) WHERE is_default = TRUE;

-- Index for active cards
CREATE INDEX idx_saved_cards_active ON saved_cards(member_id, is_active) WHERE is_active = TRUE;

-- Index for card token lookups
CREATE UNIQUE INDEX idx_saved_cards_token ON saved_cards(card_token);

-- Ensure only one default card per member
CREATE UNIQUE INDEX idx_saved_cards_one_default
    ON saved_cards(member_id)
    WHERE is_default = TRUE AND is_active = TRUE;

-- Add trigger for updating updated_at timestamp
CREATE OR REPLACE FUNCTION update_saved_card_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER saved_cards_updated_at
    BEFORE UPDATE ON saved_cards
    FOR EACH ROW
    EXECUTE FUNCTION update_saved_card_updated_at();
