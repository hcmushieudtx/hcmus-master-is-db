// Series recommendation query.
//
// Returns all active books in the same series, ordered by volume_order.
// The caller (Go layer) cross-references mongo_ids against the user's order
// history to populate the already_bought flag before returning to the client.

MATCH (b:Book {is_active: true})-[r:IN_SERIES]->(s:Series {name: $seriesName})
RETURN b.mongo_id    AS mongo_id,
       b.title       AS title,
       r.volume_order AS volume_order
ORDER BY r.volume_order ASC
