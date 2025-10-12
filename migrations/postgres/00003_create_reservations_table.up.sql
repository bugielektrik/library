CREATE TABLE IF NOT EXISTS reservations (
    id            UUID PRIMARY KEY DEFAULT GEN_RANDOM_UUID(),
    book_id       UUID NOT NULL REFERENCES books (id) ON DELETE CASCADE,
    member_id     UUID NOT NULL REFERENCES members (id) ON DELETE CASCADE,
    status        VARCHAR(20) NOT NULL DEFAULT 'pending',
    created_at    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at    TIMESTAMP NOT NULL,
    fulfilled_at  TIMESTAMP,
    cancelled_at  TIMESTAMP,
    updated_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT status_check CHECK (status IN ('pending', 'fulfilled', 'cancelled', 'expired'))
);

-- Index for faster queries by member
CREATE INDEX idx_reservations_member_id ON reservations(member_id);

-- Index for faster queries by book
CREATE INDEX idx_reservations_book_id ON reservations(book_id);

-- Index for faster queries by status
CREATE INDEX idx_reservations_status ON reservations(status);

-- Index for finding active reservations for a member and book
CREATE INDEX idx_reservations_active ON reservations(member_id, book_id, status) WHERE status IN ('pending', 'fulfilled');

-- Index for finding expired reservations
CREATE INDEX idx_reservations_expired ON reservations(expires_at) WHERE status = 'pending';
