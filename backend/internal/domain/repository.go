package domain

import (
	"context"

	"github.com/google/uuid"
)

// ─── PostgreSQL repositories ─────────────────────────────────────────────────

// UserRepository covers all user persistence operations backed by PostgreSQL.
type UserRepository interface {
	CreateUser(ctx context.Context, user *User) error
	GetUserByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	UpdateUser(ctx context.Context, user *User) error
	ListUsers(ctx context.Context, page, pageSize int) ([]*User, int64, error)
	DeactivateUser(ctx context.Context, id uuid.UUID, active bool) error
}

// OrderRepository covers order persistence operations backed by PostgreSQL.
type OrderRepository interface {
	CreateOrder(ctx context.Context, order *Order) error
	GetOrderByID(ctx context.Context, id uuid.UUID) (*Order, error)
	ListOrdersByUser(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]*Order, int64, error)
	ListAllOrders(ctx context.Context, status OrderStatus, page, pageSize int) ([]*Order, int64, error)
	UpdateOrderStatus(ctx context.Context, id uuid.UUID, status OrderStatus) error
}

// BookRefRepository manages the PostgreSQL bridge table between MongoDB book
// documents and live stock/price data.
type BookRefRepository interface {
	GetBookRef(ctx context.Context, mongoID string) (*BookRef, error)
	GetBookRefForUpdate(ctx context.Context, mongoID string) (*BookRef, error)
	CreateBookRef(ctx context.Context, ref *BookRef) error
	UpdateBookRef(ctx context.Context, ref *BookRef) error
	UpdateStock(ctx context.Context, mongoID string, delta int) error
}

// PostgresTransactor groups all PostgreSQL repositories under a single
// transaction scope. Passing a PostgresTransactor to handler code keeps
// the MongoDB/Redis/Neo4j calls outside the SQL transaction boundary.
type PostgresTransactor interface {
	UserRepository
	OrderRepository
	BookRefRepository
	// Transaction runs fn inside a single PostgreSQL ACID transaction.
	// If fn returns an error the transaction is rolled back; otherwise committed.
	Transaction(ctx context.Context, fn func(tx PostgresTransactor) error) error
}

// ─── MongoDB repository ───────────────────────────────────────────────────────

// BookFilter holds optional filter criteria for book search/listing queries.
type BookFilter struct {
	Query     string  // full-text search term (title, author)
	Genre     string
	Author    string
	Publisher string
	MinPrice  float64
	MaxPrice  float64
	Page      int
	PageSize  int
}

// BookRepository covers all book-catalog operations backed by MongoDB.
type BookRepository interface {
	SearchBooks(ctx context.Context, filter BookFilter) ([]*Book, int64, error)
	GetBookByID(ctx context.Context, id string) (*Book, error)
	GetBooksByIDs(ctx context.Context, ids []string) ([]*Book, error)
	GetNewestBooks(ctx context.Context, limit int) ([]*Book, error)
	CreateBook(ctx context.Context, book *Book) (string, error)
	UpdateBook(ctx context.Context, id string, book *Book) error
	DeleteBook(ctx context.Context, id string) error
}

// ─── Redis repositories ───────────────────────────────────────────────────────

// SessionRepository manages JWT session tokens and the blacklist in Redis.
type SessionRepository interface {
	SetToken(ctx context.Context, userID string, token string) error
	GetToken(ctx context.Context, userID string) (string, error)
	BlacklistToken(ctx context.Context, token string) error
	IsBlacklisted(ctx context.Context, token string) (bool, error)
}

// CartRepository manages shopping-cart state in a Redis Hash per user.
type CartRepository interface {
	AddItem(ctx context.Context, userID string, item CartItem) error
	GetCart(ctx context.Context, userID string) ([]CartItem, error)
	UpdateItem(ctx context.Context, userID string, bookID string, qty int) error
	RemoveItem(ctx context.Context, userID string, bookID string) error
	ClearCart(ctx context.Context, userID string) error
}

// TrendingRepository manages the Redis Sorted Set used for bestseller rankings.
type TrendingRepository interface {
	IncrScore(ctx context.Context, bookID string, delta float64) error
	GetTop10(ctx context.Context) ([]TrendingBook, error)
	SetTop10(ctx context.Context, books []TrendingBook) error
}

// ─── Neo4j repository ─────────────────────────────────────────────────────────

// RecommendationRepository issues graph traversal queries against Neo4j.
type RecommendationRepository interface {
	GetSimilarBooks(ctx context.Context, mongoID string, limit int) ([]SimilarBook, error)
	GetSeriesBooks(ctx context.Context, seriesName string) ([]SeriesBook, error)
	UpsertBookNode(ctx context.Context, node BookNode) error
	DeleteBookNode(ctx context.Context, mongoID string) error
}
