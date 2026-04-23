CREATE TYPE order_status AS ENUM (
    'pending',
    'confirmed',
    'shipping',
    'completed',
    'cancelled'
);

-- ── Orders (header) ───────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS orders (
    id               UUID         NOT NULL DEFAULT gen_random_uuid(),
    user_id          UUID         NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    status           order_status NOT NULL DEFAULT 'pending',
    total_amount     NUMERIC(14, 2) NOT NULL CHECK (total_amount >= 0),
    shipping_address TEXT         NOT NULL,
    payment_method   VARCHAR(50)  NOT NULL,
    created_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    CONSTRAINT orders_pkey PRIMARY KEY (id)
);

CREATE INDEX idx_orders_user_id   ON orders (user_id);
CREATE INDEX idx_orders_status    ON orders (status);
CREATE INDEX idx_orders_created   ON orders (created_at DESC);

-- ── Order items (lines) ───────────────────────────────────────────────────────
-- mongo_book_id and title are denormalised here so that order history remains
-- readable even if the MongoDB document is later removed or updated.
CREATE TABLE IF NOT EXISTS order_items (
    id            UUID           NOT NULL DEFAULT gen_random_uuid(),
    order_id      UUID           NOT NULL REFERENCES orders (id) ON DELETE CASCADE,
    mongo_book_id TEXT           NOT NULL,
    title         TEXT           NOT NULL,
    quantity      INTEGER        NOT NULL CHECK (quantity > 0),
    unit_price    NUMERIC(12, 2) NOT NULL CHECK (unit_price > 0),

    CONSTRAINT order_items_pkey PRIMARY KEY (id)
);

CREATE INDEX idx_order_items_order_id ON order_items (order_id);
