package middleware

import "bookstore/backend/internal/domain"

const (
	// CtxUserID is the Gin context key that holds the authenticated user's UUID string.
	CtxUserID = "userID"
	// CtxUserRole is the Gin context key that holds the authenticated user's role.
	CtxUserRole = "userRole"
	// CtxToken is the Gin context key that holds the raw JWT string for blacklist checks.
	CtxToken = "token"

	// HeaderAuthorization is the HTTP header name for Bearer tokens.
	HeaderAuthorization = "Authorization"
	// BearerPrefix is the prefix stripped from the Authorization header value.
	BearerPrefix = "Bearer "
)

// Exported role constants for use in route guards and tests.
const (
	RoleUser  = domain.RoleUser
	RoleAdmin = domain.RoleAdmin
)
