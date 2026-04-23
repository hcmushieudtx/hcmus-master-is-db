// ── Drop all indexes ──────────────────────────────────────────────────────────

DROP INDEX book_active_index IF EXISTS;
DROP INDEX book_title_index  IF EXISTS;

// ── Drop all constraints (also removes their backing indexes) ─────────────────

DROP CONSTRAINT series_name_unique    IF EXISTS;
DROP CONSTRAINT publisher_name_unique IF EXISTS;
DROP CONSTRAINT author_name_unique    IF EXISTS;
DROP CONSTRAINT genre_name_unique     IF EXISTS;
DROP CONSTRAINT book_mongo_id_unique  IF EXISTS;
