package redis

import (
	"bookstore/backend/internal/domain"
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"
)

const cartPrefix = "cart:" // cart:<userID> → Hash of bookID → JSON(CartItem)

// CartRepository implements domain.CartRepository using Redis Hashes.
// Each user's cart is stored as HSET cart:<userID> <bookID> <json>.
type CartRepository struct {
	rdb *client
}

// NewCartRepository creates a CartRepository.
func NewCartRepository(rdb *redis.Client) *CartRepository {
	return &CartRepository{rdb: rdb}
}

func cartKey(userID string) string { return cartPrefix + userID }

// AddItem adds or replaces a cart item. If the book is already in the cart
// the quantity is accumulated.
func (r *CartRepository) AddItem(ctx context.Context, userID string, item domain.CartItem) error {
	key := cartKey(userID)

	// Check if item already exists
	existing, err := r.rdb.HGet(ctx, key, item.BookID).Result()
	if err != nil && err != redis.Nil {
		return fmt.Errorf("hget cart item: %w", err)
	}

	if err == nil {
		var existingItem domain.CartItem
		if jsonErr := json.Unmarshal([]byte(existing), &existingItem); jsonErr == nil {
			item.Quantity += existingItem.Quantity
		}
	}

	data, err := json.Marshal(item)
	if err != nil {
		return fmt.Errorf("marshal cart item: %w", err)
	}
	return r.rdb.HSet(ctx, key, item.BookID, data).Err()
}

// GetCart retrieves all items in a user's cart.
func (r *CartRepository) GetCart(ctx context.Context, userID string) ([]domain.CartItem, error) {
	key := cartKey(userID)
	fields, err := r.rdb.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("hgetall cart: %w", err)
	}

	items := make([]domain.CartItem, 0, len(fields))
	for _, raw := range fields {
		var item domain.CartItem
		if err := json.Unmarshal([]byte(raw), &item); err == nil {
			items = append(items, item)
		}
	}
	return items, nil
}

// UpdateItem sets the quantity of an existing cart item.
// If qty == 0 the item is removed.
func (r *CartRepository) UpdateItem(ctx context.Context, userID string, bookID string, qty int) error {
	if qty <= 0 {
		return r.RemoveItem(ctx, userID, bookID)
	}

	key := cartKey(userID)
	raw, err := r.rdb.HGet(ctx, key, bookID).Result()
	if err == redis.Nil {
		return nil // item not in cart — nothing to update
	}
	if err != nil {
		return fmt.Errorf("hget cart item: %w", err)
	}

	var item domain.CartItem
	if err := json.Unmarshal([]byte(raw), &item); err != nil {
		return fmt.Errorf("unmarshal cart item: %w", err)
	}

	item.Quantity = qty
	data, err := json.Marshal(item)
	if err != nil {
		return fmt.Errorf("marshal cart item: %w", err)
	}
	return r.rdb.HSet(ctx, key, bookID, data).Err()
}

// RemoveItem deletes a single item from the cart.
func (r *CartRepository) RemoveItem(ctx context.Context, userID string, bookID string) error {
	return r.rdb.HDel(ctx, cartKey(userID), bookID).Err()
}

// ClearCart removes the entire cart hash for a user.
func (r *CartRepository) ClearCart(ctx context.Context, userID string) error {
	return r.rdb.Del(ctx, cartKey(userID)).Err()
}
