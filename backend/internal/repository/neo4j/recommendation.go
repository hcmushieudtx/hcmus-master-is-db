package neo4j

import (
	"bookstore/backend/internal/domain"
	"context"
	"strings"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// RecommendationRepository implements domain.RecommendationRepository against Neo4j.
type RecommendationRepository struct {
	driver neo4j.DriverWithContext
}

// NewRecommendationRepository creates a RecommendationRepository.
func NewRecommendationRepository(driver neo4j.DriverWithContext) *RecommendationRepository {
	return &RecommendationRepository{driver: driver}
}

// GetSimilarBooks traverses the book graph and returns the top similar books
// ranked by a weighted score: Genre (×3) > Author (×2) > Publisher (×1).
func (r *RecommendationRepository) GetSimilarBooks(ctx context.Context, mongoID string, limit int) ([]domain.SimilarBook, error) {
	cypher := `
MATCH (source:Book {mongo_id: $mongoID, is_active: true})

OPTIONAL MATCH (source)-[:SAME_GENRE]->(g:Genre)<-[:SAME_GENRE]-(sim:Book {is_active: true})
  WHERE sim.mongo_id <> $mongoID
WITH source, sim, COUNT(g) * 3 AS genreScore

OPTIONAL MATCH (source)-[:SAME_AUTHOR]->(a:Author)<-[:SAME_AUTHOR]-(sim)
WITH source, sim, genreScore, COUNT(a) * 2 AS authorScore

OPTIONAL MATCH (source)-[:SAME_PUBLISHER]->(p:Publisher)<-[:SAME_PUBLISHER]-(sim)
WITH sim, genreScore + authorScore + COUNT(p) AS totalScore

WHERE sim IS NOT NULL
RETURN sim.mongo_id AS mongo_id,
       sim.title    AS title,
       totalScore   AS score
ORDER BY score DESC
LIMIT $limit`

	records, err := runQuery(ctx, r.driver, cypher, map[string]any{
		"mongoID": mongoID,
		"limit":   limit,
	})
	if err != nil {
		return nil, err
	}

	result := make([]domain.SimilarBook, 0, len(records))
	for _, rec := range records {
		bookID, _ := rec.Get("mongo_id")
		title, _ := rec.Get("title")
		score, _ := rec.Get("score")

		result = append(result, domain.SimilarBook{
			BookID: asString(bookID),
			Title:  asString(title),
			Score:  asFloat64(score),
		})
	}
	return result, nil
}

// GetSeriesBooks returns all active books in a named series, ordered by volume.
func (r *RecommendationRepository) GetSeriesBooks(ctx context.Context, seriesName string) ([]domain.SeriesBook, error) {
	cypher := `
MATCH (b:Book {is_active: true})-[r:IN_SERIES]->(s:Series {name: $seriesName})
RETURN b.mongo_id    AS mongo_id,
       b.title       AS title,
       r.volume_order AS volume_order
ORDER BY r.volume_order ASC`

	records, err := runQuery(ctx, r.driver, cypher, map[string]any{"seriesName": seriesName})
	if err != nil {
		return nil, err
	}

	result := make([]domain.SeriesBook, 0, len(records))
	for _, rec := range records {
		bookID, _ := rec.Get("mongo_id")
		title, _ := rec.Get("title")
		vol, _ := rec.Get("volume_order")

		result = append(result, domain.SeriesBook{
			BookID:      asString(bookID),
			Title:       asString(title),
			VolumeOrder: asInt(vol),
		})
	}
	return result, nil
}

// UpsertBookNode creates or updates a Book node and its Genre/Author/Publisher/Series edges.
func (r *RecommendationRepository) UpsertBookNode(ctx context.Context, node domain.BookNode) error {
	cypher := `
MERGE (b:Book {mongo_id: $mongoID})
SET b.title    = $title,
    b.is_active = $isActive

WITH b
UNWIND $genres AS genreName
  MERGE (g:Genre {name: genreName})
  MERGE (b)-[:SAME_GENRE]->(g)

WITH b
UNWIND $authors AS authorName
  MERGE (a:Author {name: authorName})
  MERGE (b)-[:SAME_AUTHOR]->(a)

WITH b
MERGE (p:Publisher {name: $publisher})
MERGE (b)-[:SAME_PUBLISHER]->(p)

WITH b
FOREACH (_ IN CASE WHEN $seriesName <> '' THEN [1] ELSE [] END |
  MERGE (s:Series {name: $seriesName})
  MERGE (b)-[:IN_SERIES {volume_order: $volumeOrder}]->(s)
)`

	return writeQuery(ctx, r.driver, cypher, map[string]any{
		"mongoID":     node.MongoID,
		"title":       node.Title,
		"isActive":    node.IsActive,
		"genres":      node.Genres,
		"authors":     node.Authors,
		"publisher":   node.Publisher,
		"seriesName":  node.SeriesName,
		"volumeOrder": node.VolumeOrder,
	})
}

// DeleteBookNode marks a book node as inactive (soft-delete in the graph).
func (r *RecommendationRepository) DeleteBookNode(ctx context.Context, mongoID string) error {
	cypher := `
MATCH (b:Book {mongo_id: $mongoID})
SET b.is_active = false`

	return writeQuery(ctx, r.driver, cypher, map[string]any{"mongoID": mongoID})
}

// ─── helpers ──────────────────────────────────────────────────────────────────

func asString(v any) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return strings.TrimSpace(s)
	}
	return ""
}

func asFloat64(v any) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case int64:
		return float64(val)
	}
	return 0
}

func asInt(v any) int {
	switch val := v.(type) {
	case int64:
		return int(val)
	case float64:
		return int(val)
	}
	return 0
}
