package domain

import "github.com/google/uuid"

// ─── Auth DTOs ────────────────────────────────────────────────────────────────

type RegisterRequest struct {
	FullName string `json:"full_name" binding:"required,min=2,max=100"`
	Email    string `json:"email"     binding:"required,email"`
	Phone    string `json:"phone"     binding:"omitempty,e164"`
	Password string `json:"password"  binding:"required,min=8"`
}

type LoginRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	AccessToken string   `json:"access_token"`
	User        UserInfo `json:"user"`
}

type UserInfo struct {
	ID       uuid.UUID `json:"id"`
	FullName string    `json:"full_name"`
	Email    string    `json:"email"`
	Phone    string    `json:"phone"`
	Role     UserRole  `json:"role"`
}

// ─── Profile DTOs ─────────────────────────────────────────────────────────────

type UpdateProfileRequest struct {
	FullName    string `json:"full_name"    binding:"omitempty,min=2,max=100"`
	Phone       string `json:"phone"        binding:"omitempty,e164"`
	DefaultAddr string `json:"default_addr" binding:"omitempty,max=500"`
}

// ─── Book DTOs ────────────────────────────────────────────────────────────────

type BookListResponse struct {
	Books    []*BookDetail `json:"books"`
	Total    int64         `json:"total"`
	Page     int           `json:"page"`
	PageSize int           `json:"page_size"`
}

type CreateBookRequest struct {
	Title       string         `json:"title"       binding:"required"`
	Authors     []string       `json:"authors"     binding:"required,min=1"`
	Publisher   string         `json:"publisher"   binding:"required"`
	PublishedAt int            `json:"published_at"`
	Genres      []string       `json:"genres"      binding:"required,min=1"`
	Description string         `json:"description"`
	CoverURL    string         `json:"cover_url"`
	Language    string         `json:"language"`
	Pages       int            `json:"pages"`
	Attributes  map[string]any `json:"attributes"`
	SeriesName  string         `json:"series_name"`
	VolumeOrder int            `json:"volume_order"`
	StockQty    int            `json:"stock_qty"   binding:"required,min=0"`
	Price       float64        `json:"price"       binding:"required,gt=0"`
}

type UpdateBookRequest struct {
	Title       string         `json:"title"`
	Authors     []string       `json:"authors"`
	Publisher   string         `json:"publisher"`
	PublishedAt int            `json:"published_at"`
	Genres      []string       `json:"genres"`
	Description string         `json:"description"`
	CoverURL    string         `json:"cover_url"`
	Language    string         `json:"language"`
	Pages       int            `json:"pages"`
	Attributes  map[string]any `json:"attributes"`
	SeriesName  string         `json:"series_name"`
	VolumeOrder int            `json:"volume_order"`
	Price       float64        `json:"price"`
}

type UpdateStockRequest struct {
	StockQty int `json:"stock_qty" binding:"required,min=0"`
}

// ─── Cart DTOs ────────────────────────────────────────────────────────────────

type AddToCartRequest struct {
	BookID   string `json:"book_id"  binding:"required"`
	Quantity int    `json:"quantity" binding:"required,min=1"`
}

type UpdateCartItemRequest struct {
	Quantity int `json:"quantity" binding:"required,min=1"`
}

type CartResponse struct {
	Items      []CartItem `json:"items"`
	TotalPrice float64    `json:"total_price"`
}

// ─── Order DTOs ───────────────────────────────────────────────────────────────

type CheckoutRequest struct {
	ShippingAddress string `json:"shipping_address" binding:"required"`
	PaymentMethod   string `json:"payment_method"   binding:"required,oneof=cod bank_transfer"`
}

type OrderListResponse struct {
	Orders   []*Order `json:"orders"`
	Total    int64    `json:"total"`
	Page     int      `json:"page"`
	PageSize int      `json:"page_size"`
}

type UpdateOrderStatusRequest struct {
	Status OrderStatus `json:"status" binding:"required,oneof=pending confirmed shipping completed cancelled"`
}

// ─── Admin User DTOs ──────────────────────────────────────────────────────────

type UserListResponse struct {
	Users    []*UserInfo `json:"users"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
}

type DeactivateUserRequest struct {
	IsActive bool `json:"is_active"`
}

// ─── Recommendation DTOs ──────────────────────────────────────────────────────

type RecommendationResponse struct {
	SimilarBooks []SimilarBook `json:"similar_books"`
	SeriesBooks  []SeriesBook  `json:"series_books,omitempty"`
}

// ─── Analytics DTOs ───────────────────────────────────────────────────────────

type SalesSummary struct {
	TotalOrders  int64   `json:"total_orders"`
	TotalRevenue float64 `json:"total_revenue"`
	DateFrom     string  `json:"date_from"`
	DateTo       string  `json:"date_to"`
}
