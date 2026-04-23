package server

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// GetSimilarBooks handles GET /api/v1/books/:id/similar (NV-E1).
// Traverses the Neo4j graph and returns the top-10 similar books scored by
// shared genre (×3), author (×2), and publisher (×1).
func (s *Service) GetSimilarBooks(c *gin.Context) {
	bookID := c.Param("id")
	limit := queryInt(c, "limit", 10)

	books, err := s.recRepo.GetSimilarBooks(c.Request.Context(), bookID, limit)
	if err != nil {
		s.logger.Error("get similar books", zap.Error(err))
		respondInternalError(c, "could not fetch recommendations")
		return
	}

	respondOK(c, books)
}

// GetSeriesBooks handles GET /api/v1/books/:id/series (NV-E2).
// Returns all volumes in the same series as the given book, ordered by volume
// number.  If the caller is authenticated, already-purchased volumes are marked.
func (s *Service) GetSeriesBooks(c *gin.Context) {
	bookID := c.Param("id")
	ctx := c.Request.Context()

	// Fetch the book to retrieve its series name from MongoDB
	book, err := s.bookRepo.GetBookByID(ctx, bookID)
	if err != nil || book == nil {
		respondNotFound(c, "book not found")
		return
	}
	if book.SeriesName == "" {
		respondOK(c, []any{})
		return
	}

	seriesBooks, err := s.recRepo.GetSeriesBooks(ctx, book.SeriesName)
	if err != nil {
		s.logger.Error("get series books", zap.Error(err))
		respondInternalError(c, "could not fetch series")
		return
	}

	respondOK(c, seriesBooks)
}

// GetTrending handles GET /api/v1/trending (NV-E3).
// Returns the top-10 bestselling books from the Redis Sorted Set cache.
func (s *Service) GetTrending(c *gin.Context) {
	books, err := s.trendRepo.GetTop10(c.Request.Context())
	if err != nil {
		s.logger.Error("get trending", zap.Error(err))
		respondInternalError(c, "could not fetch trending books")
		return
	}

	respondOK(c, books)
}
