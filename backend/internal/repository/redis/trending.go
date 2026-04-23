package redis

import (
	"bookstore/backend/internal/domain"
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"
)

const (
	trendingZSet    = "trending:books"      // ZSET: member=bookID, score=sales count
	trendingCacheKey = "trending:top10"     // STRING: JSON array of TrendingBook
)

// TrendingRepository implements domain.TrendingRepository using Redis Sorted Sets.
type TrendingRepository struct {
	rdb *client
}

// NewTrendingRepository creates a TrendingRepository.
func NewTrendingRepository(rdb *redis.Client) *TrendingRepository {
	return &TrendingRepository{rdb: rdb}
}

// IncrScore increments the sales score for a book by delta.
// Called after a successful checkout for each purchased book.
func (r *TrendingRepository) IncrScore(ctx context.Context, bookID string, delta float64) error {
	return r.rdb.ZIncrBy(ctx, trendingZSet, delta, bookID).Err()
}

// GetTop10 returns the top-10 books from the pre-computed cache key.
// Falls back to computing directly from the ZSET if the cache is empty.
func (r *TrendingRepository) GetTop10(ctx context.Context) ([]domain.TrendingBook, error) {
	// Try pre-computed cache first
	raw, err := r.rdb.Get(ctx, trendingCacheKey).Result()
	if err == nil {
		var books []domain.TrendingBook
		if jsonErr := json.Unmarshal([]byte(raw), &books); jsonErr == nil {
			return books, nil
		}
	}

	// Fall back to live ZSET query
	return r.computeTop10(ctx)
}

// SetTop10 stores the pre-computed top-10 list in the cache (for the background worker).
func (r *TrendingRepository) SetTop10(ctx context.Context, books []domain.TrendingBook) error {
	data, err := json.Marshal(books)
	if err != nil {
		return fmt.Errorf("marshal trending: %w", err)
	}
	// Cache for 1 hour; background worker refreshes it periodically
	return r.rdb.Set(ctx, trendingCacheKey, data, 0).Err()
}

// computeTop10 reads directly from the ZSET sorted by score descending.
func (r *TrendingRepository) computeTop10(ctx context.Context) ([]domain.TrendingBook, error) {
	entries, err := r.rdb.ZRevRangeWithScores(ctx, trendingZSet, 0, 9).Result()
	if err != nil {
		return nil, fmt.Errorf("zrevrange trending: %w", err)
	}

	books := make([]domain.TrendingBook, 0, len(entries))
	for _, z := range entries {
		books = append(books, domain.TrendingBook{
			BookID: fmt.Sprint(z.Member),
			Score:  z.Score,
		})
	}
	return books, nil
}
