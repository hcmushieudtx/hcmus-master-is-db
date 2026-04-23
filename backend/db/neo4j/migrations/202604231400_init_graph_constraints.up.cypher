// ── Node uniqueness constraints ───────────────────────────────────────────────
// Each constraint also creates a backing index automatically.

CREATE CONSTRAINT book_mongo_id_unique IF NOT EXISTS
FOR (b:Book) REQUIRE b.mongo_id IS UNIQUE;

CREATE CONSTRAINT genre_name_unique IF NOT EXISTS
FOR (g:Genre) REQUIRE g.name IS UNIQUE;

CREATE CONSTRAINT author_name_unique IF NOT EXISTS
FOR (a:Author) REQUIRE a.name IS UNIQUE;

CREATE CONSTRAINT publisher_name_unique IF NOT EXISTS
FOR (p:Publisher) REQUIRE p.name IS UNIQUE;

CREATE CONSTRAINT series_name_unique IF NOT EXISTS
FOR (s:Series) REQUIRE s.name IS UNIQUE;

// ── Additional property indexes ───────────────────────────────────────────────

CREATE INDEX book_title_index IF NOT EXISTS
FOR (b:Book) ON (b.title);

CREATE INDEX book_active_index IF NOT EXISTS
FOR (b:Book) ON (b.is_active);

// ── Relationship types (documentation only — Neo4j relationships are schema-free)
//
// (:Book)-[:SAME_GENRE    {weight: 3}]->(:Genre)
// (:Book)-[:SAME_AUTHOR   {weight: 2}]->(:Author)
// (:Book)-[:SAME_PUBLISHER{weight: 1}]->(:Publisher)
// (:Book)-[:IN_SERIES     {volume_order: <int>}]->(:Series)
