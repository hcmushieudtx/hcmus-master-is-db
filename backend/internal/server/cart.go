package server

import (
	"bookstore/backend/internal/domain"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AddToCart handles POST /api/v1/cart (NV-C1).
func (s *Service) AddToCart(c *gin.Context) {
	var req domain.AddToCartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBadRequest(c, err.Error())
		return
	}

	ctx := c.Request.Context()
	userID := mustUserID(c).String()

	// Verify the book exists and has sufficient stock
	ref, err := s.pg.GetBookRef(ctx, req.BookID)
	if err != nil || ref == nil || !ref.IsActive {
		respondNotFound(c, "book not found or unavailable")
		return
	}
	if ref.StockQty < req.Quantity {
		respondBadRequest(c, "insufficient stock")
		return
	}

	// Fetch title from MongoDB for display purposes
	title := req.BookID
	if book, err := s.bookRepo.GetBookByID(ctx, req.BookID); err == nil && book != nil {
		title = book.Title
	}

	item := domain.CartItem{
		BookID:   req.BookID,
		Title:    title,
		Price:    ref.Price,
		Quantity: req.Quantity,
	}

	if err := s.cartRepo.AddItem(ctx, userID, item); err != nil {
		s.logger.Error("add to cart", zap.Error(err))
		respondInternalError(c, "could not add item to cart")
		return
	}

	respondOK(c, gin.H{"message": "item added to cart"})
}

// GetCart handles GET /api/v1/cart (NV-C2).
func (s *Service) GetCart(c *gin.Context) {
	userID := mustUserID(c).String()
	items, err := s.cartRepo.GetCart(c.Request.Context(), userID)
	if err != nil {
		s.logger.Error("get cart", zap.Error(err))
		respondInternalError(c, "could not retrieve cart")
		return
	}

	var total float64
	for _, it := range items {
		total += it.Price * float64(it.Quantity)
	}

	respondOK(c, domain.CartResponse{Items: items, TotalPrice: total})
}

// UpdateCartItem handles PUT /api/v1/cart/:bookId (NV-C2).
func (s *Service) UpdateCartItem(c *gin.Context) {
	var req domain.UpdateCartItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBadRequest(c, err.Error())
		return
	}

	bookID := c.Param("bookId")
	userID := mustUserID(c).String()

	if err := s.cartRepo.UpdateItem(c.Request.Context(), userID, bookID, req.Quantity); err != nil {
		s.logger.Error("update cart item", zap.Error(err))
		respondInternalError(c, "could not update cart item")
		return
	}

	respondOK(c, gin.H{"message": "cart item updated"})
}

// RemoveCartItem handles DELETE /api/v1/cart/:bookId (NV-C2).
func (s *Service) RemoveCartItem(c *gin.Context) {
	bookID := c.Param("bookId")
	userID := mustUserID(c).String()

	if err := s.cartRepo.RemoveItem(c.Request.Context(), userID, bookID); err != nil {
		s.logger.Error("remove cart item", zap.Error(err))
		respondInternalError(c, "could not remove cart item")
		return
	}

	respondOK(c, gin.H{"message": "item removed from cart"})
}
