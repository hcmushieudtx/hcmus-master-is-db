package postgres

import (
	"bookstore/backend/internal/domain"
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CreateUser inserts a new user row.
func (q *Queries) CreateUser(ctx context.Context, user *domain.User) error {
	return q.db.WithContext(ctx).Create(user).Error
}

// GetUserByID fetches a user by primary key.
func (q *Queries) GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	var user domain.User
	err := q.db.WithContext(ctx).First(&user, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &user, err
}

// GetUserByEmail fetches a user by their unique email address.
func (q *Queries) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	err := q.db.WithContext(ctx).First(&user, "email = ?", email).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &user, err
}

// UpdateUser persists changes to full_name, phone, and default_addr.
func (q *Queries) UpdateUser(ctx context.Context, user *domain.User) error {
	return q.db.WithContext(ctx).Model(user).
		Select("full_name", "phone", "default_addr").
		Updates(user).Error
}

// ListUsers returns a paginated list of all users ordered by creation time.
func (q *Queries) ListUsers(ctx context.Context, page, pageSize int) ([]*domain.User, int64, error) {
	var users []*domain.User
	var total int64

	if err := q.db.WithContext(ctx).Model(&domain.User{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := q.db.WithContext(ctx).
		Order("created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&users).Error

	return users, total, err
}

// DeactivateUser toggles the is_active flag for a given user.
func (q *Queries) DeactivateUser(ctx context.Context, id uuid.UUID, active bool) error {
	return q.db.WithContext(ctx).
		Model(&domain.User{}).
		Where("id = ?", id).
		Update("is_active", active).Error
}
