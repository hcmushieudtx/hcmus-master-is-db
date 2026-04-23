package server

import (
	"bookstore/backend/internal/domain"
	"context"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// SearchBooks handles GET /api/v1/books (NV-B1).
// Supports query params: q, genre, author, publisher, page, page_size.
func (s *Service) SearchBooks(c *gin.Context) {
	filter := domain.BookFilter{
		Query:     c.Query("q"),
		Genre:     c.Query("genre"),
		Author:    c.Query("author"),
		Publisher: c.Query("publisher"),
		Page:      queryInt(c, "page", 1),
		PageSize:  queryInt(c, "page_size", 20),
	}

	ctx := c.Request.Context()
	books, total, err := s.bookRepo.SearchBooks(ctx, filter)
	if err != nil {
		s.logger.Error("search books", zap.Error(err))
		respondInternalError(c, "could not search books")
		return
	}

	details := s.enrichBooks(ctx, books)
	respondPaginated(c, details, total, filter.Page, filter.PageSize)
}

// GetBookDetail handles GET /api/v1/books/:id (NV-B2).
func (s *Service) GetBookDetail(c *gin.Context) {
	bookID := c.Param("id")
	ctx := c.Request.Context()

	book, err := s.bookRepo.GetBookByID(ctx, bookID)
	if err != nil || book == nil {
		respondNotFound(c, "book not found")
		return
	}

	ref, _ := s.pg.GetBookRef(ctx, bookID)
	detail := domain.BookDetail{Book: *book}
	if ref != nil {
		detail.StockQty = ref.StockQty
		detail.Price = ref.Price
	}

	respondOK(c, detail)
}

// GetNewBooks handles GET /api/v1/books/new (NV-B3).
func (s *Service) GetNewBooks(c *gin.Context) {
	limit := queryInt(c, "limit", 20)
	ctx := c.Request.Context()

	books, err := s.bookRepo.GetNewestBooks(ctx, limit)
	if err != nil {
		s.logger.Error("get newest books", zap.Error(err))
		respondInternalError(c, "could not fetch new books")
		return
	}

	respondOK(c, s.enrichBooks(ctx, books))
}

// ─── helpers ──────────────────────────────────────────────────────────────────

// enrichBooks fetches BookRef rows from PostgreSQL and merges stock/price data
// into each book.  Missing refs result in zero-value stock and price.
func (s *Service) enrichBooks(ctx context.Context, books []*domain.Book) []domain.BookDetail {
	details := make([]domain.BookDetail, 0, len(books))
	for _, b := range books {
		detail := domain.BookDetail{Book: *b}
		if ref, err := s.pg.GetBookRef(ctx, b.ID); err == nil && ref != nil {
			detail.StockQty = ref.StockQty
			detail.Price = ref.Price
		}
		details = append(details, detail)
	}
	return details
}

// queryInt reads a query parameter as int, falling back to the provided default.
func queryInt(c *gin.Context, key string, fallback int) int {
	v, err := strconv.Atoi(c.Query(key))
	if err != nil || v < 1 {
		return fallback
	}
	return v
}
