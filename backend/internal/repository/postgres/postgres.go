package postgres

import (
	"bookstore/backend/internal/domain"
	"context"

	"gorm.io/gorm"
)

// Queries is the PostgreSQL adapter that implements domain.PostgresTransactor.
// It wraps a GORM DB handle and provides transaction support.
type Queries struct {
	db *gorm.DB
}

// New creates a Queries instance backed by the given GORM DB.
func New(db *gorm.DB) *Queries {
	return &Queries{db: db}
}

// withDB returns a new Queries scoped to the given DB (used internally for transactions).
func (q *Queries) withDB(db *gorm.DB) *Queries {
	return &Queries{db: db}
}

// Ping verifies the underlying database connection is alive.
func (q *Queries) Ping() error {
	sqlDB, err := q.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}

// Transaction executes fn inside a single ACID PostgreSQL transaction.
// The PostgresTransactor passed to fn is scoped to that transaction so all
// repository calls within fn share the same *sql.Tx.
func (q *Queries) Transaction(ctx context.Context, fn func(tx domain.PostgresTransactor) error) error {
	return q.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(q.withDB(tx))
	})
}
