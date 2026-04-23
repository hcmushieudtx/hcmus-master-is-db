-- books_ref bridges MongoDB book documents (identified by mongo_id) with
-- PostgreSQL-managed stock quantities and prices.  Keeping these fields in
-- PostgreSQL lets us use SELECT … FOR UPDATE for race-free stock deduction
-- during checkout without touching MongoDB inside a SQL transaction.

CREATE TABLE IF NOT EXISTS books_ref (
    id        UUID    NOT NULL DEFAULT gen_random_uuid(),
    mongo_id  TEXT    NOT NULL,
    stock_qty INTEGER NOT NULL DEFAULT 0 CHECK (stock_qty >= 0),
    price     NUMERIC(12, 2) NOT NULL CHECK (price > 0),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,

    CONSTRAINT books_ref_pkey          PRIMARY KEY (id),
    CONSTRAINT books_ref_mongo_id_uq   UNIQUE      (mongo_id)
);

CREATE INDEX idx_books_ref_mongo_id  ON books_ref (mongo_id);
CREATE INDEX idx_books_ref_is_active ON books_ref (is_active);
