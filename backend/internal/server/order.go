package server

import (
	"bookstore/backend/internal/domain"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Checkout handles POST /api/v1/orders/checkout (NV-D1).
// This is the critical atomic flow:
//  1. Load cart from Redis
//  2. Begin PostgreSQL transaction
//  3. Lock each books_ref row (SELECT FOR UPDATE) — verify stock
//  4. Create order header
//  5. Insert order_items (with price snapshot from Redis cart)
//  6. Decrement stock for each item
//  7. Commit
//  8. Clear Redis cart
//  9. Update trending scores in Redis
func (s *Service) Checkout(c *gin.Context) {
	var req domain.CheckoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBadRequest(c, err.Error())
		return
	}

	ctx := c.Request.Context()
	userID := mustUserID(c)

	// Step 1: Load cart
	cartItems, err := s.cartRepo.GetCart(ctx, userID.String())
	if err != nil || len(cartItems) == 0 {
		respondBadRequest(c, "cart is empty")
		return
	}

	var createdOrder *domain.Order

	// Steps 2–7: Single PostgreSQL transaction
	txErr := s.pg.Transaction(ctx, func(tx domain.PostgresTransactor) error {
		var total float64
		items := make([]domain.OrderItem, 0, len(cartItems))

		for _, ci := range cartItems {
			// Step 3: Lock row and check stock
			ref, err := tx.GetBookRefForUpdate(ctx, ci.BookID)
			if err != nil || ref == nil {
				return fmt.Errorf("book %s not found", ci.BookID)
			}
			if !ref.IsActive {
				return fmt.Errorf("book %s is no longer available", ci.BookID)
			}
			if ref.StockQty < ci.Quantity {
				return fmt.Errorf("insufficient stock for book %s", ci.BookID)
			}

			items = append(items, domain.OrderItem{
				MongoBookID: ci.BookID,
				Title:       ci.Title,
				Quantity:    ci.Quantity,
				UnitPrice:   ci.Price, // snapshot from cart (set at add-to-cart time)
			})
			total += ci.Price * float64(ci.Quantity)

			// Step 6: Decrement stock
			if err := tx.UpdateStock(ctx, ci.BookID, -ci.Quantity); err != nil {
				return fmt.Errorf("update stock for %s: %w", ci.BookID, err)
			}
		}

		// Steps 4 & 5: Create order
		order := &domain.Order{
			UserID:          userID,
			Status:          domain.OrderStatusPending,
			TotalAmount:     total,
			ShippingAddress: req.ShippingAddress,
			PaymentMethod:   req.PaymentMethod,
			Items:           items,
		}

		if err := tx.CreateOrder(ctx, order); err != nil {
			return fmt.Errorf("create order: %w", err)
		}
		createdOrder = order
		return nil
	})

	if txErr != nil {
		s.logger.Warn("checkout transaction failed", zap.Error(txErr))
		respondError(c, http.StatusConflict, txErr.Error())
		return
	}

	// Step 8: Clear cart (outside TX — non-critical if this fails)
	if err := s.cartRepo.ClearCart(ctx, userID.String()); err != nil {
		s.logger.Warn("clear cart after checkout", zap.Error(err))
	}

	// Step 9: Update Redis trending scores
	for _, ci := range cartItems {
		if err := s.trendRepo.IncrScore(ctx, ci.BookID, float64(ci.Quantity)); err != nil {
			s.logger.Warn("incr trending score", zap.String("bookID", ci.BookID), zap.Error(err))
		}
	}

	respondCreated(c, createdOrder)
}

// GetOrderHistory handles GET /api/v1/orders (NV-D2).
func (s *Service) GetOrderHistory(c *gin.Context) {
	userID := mustUserID(c)
	page := queryInt(c, "page", 1)
	pageSize := queryInt(c, "page_size", 10)

	orders, total, err := s.pg.ListOrdersByUser(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		s.logger.Error("list orders", zap.Error(err))
		respondInternalError(c, "could not fetch orders")
		return
	}

	respondPaginated(c, orders, total, page, pageSize)
}

// GetOrderDetail handles GET /api/v1/orders/:id (NV-D3).
func (s *Service) GetOrderDetail(c *gin.Context) {
	orderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "invalid order id")
		return
	}

	userID := mustUserID(c)
	order, err := s.pg.GetOrderByID(c.Request.Context(), orderID)
	if err != nil || order == nil {
		respondNotFound(c, "order not found")
		return
	}

	// Customers may only view their own orders
	if order.UserID != userID {
		respondForbidden(c, "access denied")
		return
	}

	respondOK(c, order)
}
