-- Enable the pgcrypto extension so gen_random_uuid() is available on PostgreSQL < 13.
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ── User roles ───────────────────────────────────────────────────────────────
CREATE TYPE user_role AS ENUM ('user', 'admin');

-- ── Users ────────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS users (
    id            UUID         NOT NULL DEFAULT gen_random_uuid(),
    full_name     VARCHAR(100) NOT NULL,
    email         VARCHAR(255) NOT NULL,
    phone         VARCHAR(20),
    password_hash TEXT         NOT NULL,
    role          user_role    NOT NULL DEFAULT 'user',
    is_active     BOOLEAN      NOT NULL DEFAULT TRUE,
    default_addr  TEXT,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    CONSTRAINT users_pkey         PRIMARY KEY (id),
    CONSTRAINT users_email_unique UNIQUE      (email)
);

CREATE INDEX idx_users_email    ON users (email);
CREATE INDEX idx_users_role     ON users (role);
CREATE INDEX idx_users_active   ON users (is_active);
