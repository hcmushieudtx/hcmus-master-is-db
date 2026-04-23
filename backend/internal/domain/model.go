package domain

import (
	"time"

	"github.com/google/uuid"
)

// ─── User ────────────────────────────────────────────────────────────────────

type UserRole string

const (
	RoleUser  UserRole = "user"
	RoleAdmin UserRole = "admin"
)

// User is the PostgreSQL-backed entity for authentication and profile data.
type User struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	FullName     string    `gorm:"not null"`
	Email        string    `gorm:"uniqueIndex;not null"`
	Phone        string
	PasswordHash string   `gorm:"not null"`
	Role         UserRole `gorm:"type:varchar(10);not null;default:'user'"`
	IsActive     bool     `gorm:"not null;default:true"`
	DefaultAddr  string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// ─── Book (Catalog — MongoDB) ─────────────────────────────────────────────────

// Book represents a book document stored in MongoDB.
// The schema is intentionally flexible to accommodate different physical formats
// (hardcover, paperback, boxset, signed editions, etc.).
type Book struct {
	ID          string         `bson:"_id,omitempty"    json:"id"`
	Title       string         `bson:"title"            json:"title"`
	Authors     []string       `bson:"authors"          json:"authors"`
	Publisher   string         `bson:"publisher"        json:"publisher"`
	PublishedAt int            `bson:"published_at"     json:"published_at"`
	Genres      []string       `bson:"genres"           json:"genres"`
	Description string         `bson:"description"      json:"description"`
	CoverURL    string         `bson:"cover_url"        json:"cover_url"`
	Language    string         `bson:"language"         json:"language"`
	Pages       int            `bson:"pages"            json:"pages"`
	// Format-specific attributes stored as a flexible map (e.g. "binding": "hardcover",
	// "dimensions": "24x16cm", "edition": "Signed", "set_volumes": 7).
	Attributes  map[string]any `bson:"attributes"       json:"attributes,omitempty"`
	SeriesName  string         `bson:"series_name"      json:"series_name,omitempty"`
	VolumeOrder int            `bson:"volume_order"     json:"volume_order,omitempty"`
	CreatedAt   time.Time      `bson:"created_at"       json:"created_at"`
	UpdatedAt   time.Time      `bson:"updated_at"       json:"updated_at"`
}

// BookRef is the PostgreSQL row that bridges MongoDB documents to stock and price data.
type BookRef struct {
	ID       uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	MongoID  string    `gorm:"uniqueIndex;not null"`
	StockQty int       `gorm:"not null;default:0"`
	Price    float64   `gorm:"not null"`
	IsActive bool      `gorm:"not null;default:true"`
}

// BookDetail combines a MongoDB Book document with live stock/price from PostgreSQL.
type BookDetail struct {
	Book
	StockQty int     `json:"stock_qty"`
	Price    float64 `json:"price"`
}

// ─── Cart (Redis) ─────────────────────────────────────────────────────────────

// CartItem represents a single line in a user's shopping cart (stored in Redis Hash).
type CartItem struct {
	BookID   string  `json:"book_id"`
	Title    string  `json:"title"`
	Price    float64 `json:"price"`
	Quantity int     `json:"quantity"`
}

// ─── Order (PostgreSQL) ───────────────────────────────────────────────────────

type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusConfirmed OrderStatus = "confirmed"
	OrderStatusShipping  OrderStatus = "shipping"
	OrderStatusCompleted OrderStatus = "completed"
	OrderStatusCancelled OrderStatus = "cancelled"
)

// Order is the header record of a placed order.
type Order struct {
	ID              uuid.UUID   `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID          uuid.UUID   `gorm:"type:uuid;not null;index"`
	Status          OrderStatus `gorm:"type:varchar(20);not null;default:'pending'"`
	TotalAmount     float64     `gorm:"not null"`
	ShippingAddress string      `gorm:"not null"`
	PaymentMethod   string      `gorm:"not null"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
	Items           []OrderItem `gorm:"foreignKey:OrderID"`
}

// OrderItem stores the price snapshot at the time of purchase so that later
// price changes in MongoDB do not retroactively alter order totals.
type OrderItem struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	OrderID    uuid.UUID `gorm:"type:uuid;not null;index"`
	MongoBookID string   `gorm:"not null"`
	Title      string    `gorm:"not null"`
	Quantity   int       `gorm:"not null"`
	UnitPrice  float64   `gorm:"not null"`
}

// ─── Neo4j Graph Nodes ────────────────────────────────────────────────────────

// BookNode is the graph representation of a book used by the recommendation engine.
type BookNode struct {
	MongoID     string   `json:"mongo_id"`
	Title       string   `json:"title"`
	Authors     []string `json:"authors"`
	Genres      []string `json:"genres"`
	Publisher   string   `json:"publisher"`
	SeriesName  string   `json:"series_name,omitempty"`
	VolumeOrder int      `json:"volume_order,omitempty"`
	IsActive    bool     `json:"is_active"`
}

// SimilarBook is a recommendation result with a computed similarity score.
type SimilarBook struct {
	BookID    string  `json:"book_id"`
	Title     string  `json:"title"`
	Score     float64 `json:"score"`
	CoverURL  string  `json:"cover_url,omitempty"`
}

// SeriesBook is a recommendation result for series/volume suggestions.
type SeriesBook struct {
	BookID      string `json:"book_id"`
	Title       string `json:"title"`
	VolumeOrder int    `json:"volume_order"`
	AlreadyBought bool `json:"already_bought"`
	CoverURL    string `json:"cover_url,omitempty"`
}

// ─── Trending (Redis) ─────────────────────────────────────────────────────────

// TrendingBook is an entry from the Redis sorted set of bestsellers.
type TrendingBook struct {
	BookID string  `json:"book_id"`
	Title  string  `json:"title"`
	Score  float64 `json:"score"`
}
