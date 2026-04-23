// Similar-book recommendation query.
//
// Strategy: traverse from book $mongoID outward through Genre, Author, and
// Publisher nodes.  Each shared node type contributes a weighted score:
//   - same Genre:     weight 3
//   - same Author:    weight 2
//   - same Publisher: weight 1
//
// The top $limit books by total score are returned (excluding the source book).

MATCH (source:Book {mongo_id: $mongoID, is_active: true})

// Collect neighbours via weighted relationships
OPTIONAL MATCH (source)-[:SAME_GENRE]->(g:Genre)<-[:SAME_GENRE]-(similar:Book {is_active: true})
  WHERE similar.mongo_id <> $mongoID
WITH source, similar, COUNT(g) * 3 AS genreScore

OPTIONAL MATCH (source)-[:SAME_AUTHOR]->(a:Author)<-[:SAME_AUTHOR]-(similar)
WITH source, similar, genreScore, COUNT(a) * 2 AS authorScore

OPTIONAL MATCH (source)-[:SAME_PUBLISHER]->(p:Publisher)<-[:SAME_PUBLISHER]-(similar)
WITH similar, genreScore + authorScore + COUNT(p) AS totalScore

WHERE similar IS NOT NULL
RETURN similar.mongo_id  AS mongo_id,
       similar.title     AS title,
       totalScore        AS score
ORDER BY score DESC
LIMIT $limit
