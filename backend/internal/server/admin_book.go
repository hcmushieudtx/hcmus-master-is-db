package server

import (
	"bookstore/backend/internal/domain"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AdminListBooks handles GET /api/v1/admin/books.
// Returns paginated books with live stock and price data.
func (s *Service) AdminListBooks(c *gin.Context) {
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
		s.logger.Error("admin list books", zap.Error(err))
		respondInternalError(c, "could not list books")
		return
	}

	details := s.enrichBooks(ctx, books)
	respondPaginated(c, details, total, filter.Page, filter.PageSize)
}

// AdminCreateBook handles POST /api/v1/admin/books.
// Orchestrates writes to MongoDB (catalog), PostgreSQL (books_ref), and Neo4j (graph node).
func (s *Service) AdminCreateBook(c *gin.Context) {
	var req domain.CreateBookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBadRequest(c, err.Error())
		return
	}

	ctx := c.Request.Context()

	// 1. Insert book document into MongoDB
	book := &domain.Book{
		Title:       req.Title,
		Authors:     req.Authors,
		Publisher:   req.Publisher,
		PublishedAt: req.PublishedAt,
		Genres:      req.Genres,
		Description: req.Description,
		CoverURL:    req.CoverURL,
		Language:    req.Language,
		Pages:       req.Pages,
		Attributes:  req.Attributes,
		SeriesName:  req.SeriesName,
		VolumeOrder: req.VolumeOrder,
	}

	mongoID, err := s.bookRepo.CreateBook(ctx, book)
	if err != nil {
		s.logger.Error("create book in mongo", zap.Error(err))
		respondInternalError(c, "could not create book")
		return
	}

	// 2. Insert bridge row into PostgreSQL
	ref := &domain.BookRef{
		MongoID:  mongoID,
		StockQty: req.StockQty,
		Price:    req.Price,
		IsActive: true,
	}
	if err := s.pg.CreateBookRef(ctx, ref); err != nil {
		s.logger.Error("create book ref in postgres", zap.Error(err))
		// Attempt rollback of MongoDB document — best effort
		_ = s.bookRepo.DeleteBook(ctx, mongoID)
		respondInternalError(c, "could not create book reference")
		return
	}

	// 3. Upsert Book node in Neo4j
	node := domain.BookNode{
		MongoID:     mongoID,
		Title:       req.Title,
		Authors:     req.Authors,
		Genres:      req.Genres,
		Publisher:   req.Publisher,
		SeriesName:  req.SeriesName,
		VolumeOrder: req.VolumeOrder,
		IsActive:    true,
	}
	if err := s.recRepo.UpsertBookNode(ctx, node); err != nil {
		s.logger.Warn("upsert neo4j node (non-fatal)", zap.Error(err))
	}

	book.ID = mongoID
	respondCreated(c, domain.BookDetail{
		Book:     *book,
		StockQty: req.StockQty,
		Price:    req.Price,
	})
}

// AdminUpdateBook handles PUT /api/v1/admin/books/:id.
func (s *Service) AdminUpdateBook(c *gin.Context) {
	bookID := c.Param("id")
	var req domain.UpdateBookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBadRequest(c, err.Error())
		return
	}

	ctx := c.Request.Context()

	existing, err := s.bookRepo.GetBookByID(ctx, bookID)
	if err != nil || existing == nil {
		respondNotFound(c, "book not found")
		return
	}

	updated := &domain.Book{
		Title:       orString(req.Title, existing.Title),
		Authors:     orStrings(req.Authors, existing.Authors),
		Publisher:   orString(req.Publisher, existing.Publisher),
		PublishedAt: orInt(req.PublishedAt, existing.PublishedAt),
		Genres:      orStrings(req.Genres, existing.Genres),
		Description: orString(req.Description, existing.Description),
		CoverURL:    orString(req.CoverURL, existing.CoverURL),
		Language:    orString(req.Language, existing.Language),
		Pages:       orInt(req.Pages, existing.Pages),
		Attributes:  req.Attributes,
		SeriesName:  orString(req.SeriesName, existing.SeriesName),
		VolumeOrder: orInt(req.VolumeOrder, existing.VolumeOrder),
	}

	if err := s.bookRepo.UpdateBook(ctx, bookID, updated); err != nil {
		s.logger.Error("update book", zap.Error(err))
		respondInternalError(c, "could not update book")
		return
	}

	// If price changed, update books_ref
	if req.Price > 0 {
		ref, _ := s.pg.GetBookRef(ctx, bookID)
		if ref != nil {
			ref.Price = req.Price
			_ = s.pg.UpdateBookRef(ctx, ref)
		}
	}

	// Keep Neo4j node in sync
	_ = s.recRepo.UpsertBookNode(ctx, domain.BookNode{
		MongoID:     bookID,
		Title:       updated.Title,
		Authors:     updated.Authors,
		Genres:      updated.Genres,
		Publisher:   updated.Publisher,
		SeriesName:  updated.SeriesName,
		VolumeOrder: updated.VolumeOrder,
		IsActive:    true,
	})

	respondOK(c, gin.H{"message": "book updated"})
}

// AdminDeleteBook handles DELETE /api/v1/admin/books/:id.
// Performs a soft-delete: marks is_active=false in PostgreSQL and Neo4j.
// The MongoDB document is retained for order history integrity.
func (s *Service) AdminDeleteBook(c *gin.Context) {
	bookID := c.Param("id")
	ctx := c.Request.Context()

	ref, err := s.pg.GetBookRef(ctx, bookID)
	if err != nil || ref == nil {
		respondNotFound(c, "book not found")
		return
	}

	ref.IsActive = false
	if err := s.pg.UpdateBookRef(ctx, ref); err != nil {
		s.logger.Error("soft delete book ref", zap.Error(err))
		respondInternalError(c, "could not delete book")
		return
	}

	_ = s.recRepo.DeleteBookNode(ctx, bookID)

	respondOK(c, gin.H{"message": "book deactivated"})
}

// AdminUpdateStock handles PATCH /api/v1/admin/books/:id/stock.
func (s *Service) AdminUpdateStock(c *gin.Context) {
	bookID := c.Param("id")
	var req domain.UpdateStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBadRequest(c, err.Error())
		return
	}

	ctx := c.Request.Context()
	ref, err := s.pg.GetBookRef(ctx, bookID)
	if err != nil || ref == nil {
		respondNotFound(c, "book not found")
		return
	}

	ref.StockQty = req.StockQty
	if err := s.pg.UpdateBookRef(ctx, ref); err != nil {
		s.logger.Error("update stock", zap.Error(err))
		respondError(c, http.StatusInternalServerError, "could not update stock")
		return
	}

	respondOK(c, gin.H{"stock_qty": req.StockQty})
}

// ─── field-level merge helpers ────────────────────────────────────────────────

func orString(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

func orStrings(a, b []string) []string {
	if len(a) > 0 {
		return a
	}
	return b
}

func orInt(a, b int) int {
	if a != 0 {
		return a
	}
	return b
}
