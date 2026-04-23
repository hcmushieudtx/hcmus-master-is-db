package postgres

import (
	"bookstore/backend/internal/domain"
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CreateOrder inserts the order header and all its line items.
// Must be called inside a Transaction to guarantee atomicity with stock updates.
func (q *Queries) CreateOrder(ctx context.Context, order *domain.Order) error {
	return q.db.WithContext(ctx).
		Omit("Items.*").
		Create(order).Error
}

// GetOrderByID fetches an order together with its line items.
func (q *Queries) GetOrderByID(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
	var order domain.Order
	err := q.db.WithContext(ctx).
		Preload("Items").
		First(&order, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &order, err
}

// ListOrdersByUser returns a paginated list of orders belonging to a single user.
func (q *Queries) ListOrdersByUser(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]*domain.Order, int64, error) {
	var orders []*domain.Order
	var total int64

	q.db.WithContext(ctx).Model(&domain.Order{}).Where("user_id = ?", userID).Count(&total) //nolint:errcheck

	offset := (page - 1) * pageSize
	err := q.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&orders).Error

	return orders, total, err
}

// ListAllOrders returns a paginated list of all orders, optionally filtered by status.
func (q *Queries) ListAllOrders(ctx context.Context, status domain.OrderStatus, page, pageSize int) ([]*domain.Order, int64, error) {
	var orders []*domain.Order
	var total int64

	tx := q.db.WithContext(ctx).Model(&domain.Order{})
	if status != "" {
		tx = tx.Where("status = ?", status)
	}
	tx.Count(&total) //nolint:errcheck

	offset := (page - 1) * pageSize
	err := tx.Order("created_at DESC").Limit(pageSize).Offset(offset).Find(&orders).Error

	return orders, total, err
}

// UpdateOrderStatus updates the status field of an existing order.
func (q *Queries) UpdateOrderStatus(ctx context.Context, id uuid.UUID, status domain.OrderStatus) error {
	return q.db.WithContext(ctx).
		Model(&domain.Order{}).
		Where("id = ?", id).
		Update("status", status).Error
}

// ── BookRef ───────────────────────────────────────────────────────────────────

// GetBookRef fetches the stock/price reference row for a MongoDB book ID.
func (q *Queries) GetBookRef(ctx context.Context, mongoID string) (*domain.BookRef, error) {
	var ref domain.BookRef
	err := q.db.WithContext(ctx).First(&ref, "mongo_id = ?", mongoID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &ref, err
}

// GetBookRefForUpdate fetches the reference row with a row-level lock (SELECT FOR UPDATE).
// Must be called inside a Transaction.
func (q *Queries) GetBookRefForUpdate(ctx context.Context, mongoID string) (*domain.BookRef, error) {
	var ref domain.BookRef
	err := q.db.WithContext(ctx).
		Set("gorm:query_option", "FOR UPDATE").
		First(&ref, "mongo_id = ?", mongoID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &ref, err
}

// CreateBookRef inserts a new bridge row.
func (q *Queries) CreateBookRef(ctx context.Context, ref *domain.BookRef) error {
	return q.db.WithContext(ctx).Create(ref).Error
}

// UpdateBookRef persists changes to stock_qty, price, and is_active.
func (q *Queries) UpdateBookRef(ctx context.Context, ref *domain.BookRef) error {
	return q.db.WithContext(ctx).Model(ref).
		Select("stock_qty", "price", "is_active").
		Updates(ref).Error
}

// UpdateStock adjusts stock_qty by delta (positive = add, negative = deduct).
// Enforces the CHECK constraint: if the result would go below 0, the DB rejects it.
func (q *Queries) UpdateStock(ctx context.Context, mongoID string, delta int) error {
	return q.db.WithContext(ctx).
		Model(&domain.BookRef{}).
		Where("mongo_id = ?", mongoID).
		UpdateColumn("stock_qty", gorm.Expr("stock_qty + ?", delta)).Error
}
