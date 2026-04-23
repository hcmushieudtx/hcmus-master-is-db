package server

import (
	"bookstore/backend/internal/domain"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// AdminListOrders handles GET /api/v1/admin/orders.
// Supports optional filter: status, page, page_size.
func (s *Service) AdminListOrders(c *gin.Context) {
	status := domain.OrderStatus(c.Query("status"))
	page := queryInt(c, "page", 1)
	pageSize := queryInt(c, "page_size", 20)

	orders, total, err := s.pg.ListAllOrders(c.Request.Context(), status, page, pageSize)
	if err != nil {
		s.logger.Error("admin list orders", zap.Error(err))
		respondInternalError(c, "could not list orders")
		return
	}

	respondPaginated(c, orders, total, page, pageSize)
}

// AdminGetOrder handles GET /api/v1/admin/orders/:id.
func (s *Service) AdminGetOrder(c *gin.Context) {
	orderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "invalid order id")
		return
	}

	order, err := s.pg.GetOrderByID(c.Request.Context(), orderID)
	if err != nil || order == nil {
		respondNotFound(c, "order not found")
		return
	}

	respondOK(c, order)
}

// AdminUpdateOrderStatus handles PATCH /api/v1/admin/orders/:id/status.
func (s *Service) AdminUpdateOrderStatus(c *gin.Context) {
	orderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "invalid order id")
		return
	}

	var req domain.UpdateOrderStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBadRequest(c, err.Error())
		return
	}

	ctx := c.Request.Context()
	order, err := s.pg.GetOrderByID(ctx, orderID)
	if err != nil || order == nil {
		respondNotFound(c, "order not found")
		return
	}

	if err := s.pg.UpdateOrderStatus(ctx, orderID, req.Status); err != nil {
		s.logger.Error("update order status", zap.Error(err))
		respondInternalError(c, "could not update order status")
		return
	}

	respondOK(c, gin.H{"status": req.Status})
}
