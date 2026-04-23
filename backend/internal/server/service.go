package server

import (
	"bookstore/backend/config"
	"bookstore/backend/internal/domain"

	"go.uber.org/zap"
)

// Service holds all repository dependencies and is shared by every handler.
// Handlers are methods on *Service, keeping wiring centralised.
type Service struct {
	pg          domain.PostgresTransactor
	bookRepo    domain.BookRepository
	recRepo     domain.RecommendationRepository
	sessionRepo domain.SessionRepository
	cartRepo    domain.CartRepository
	trendRepo   domain.TrendingRepository
	jwtCfg      config.JWTConfig
	logger      *zap.Logger
}

// NewService creates a Service with all dependencies injected.
func NewService(
	pg domain.PostgresTransactor,
	bookRepo domain.BookRepository,
	recRepo domain.RecommendationRepository,
	sessionRepo domain.SessionRepository,
	cartRepo domain.CartRepository,
	trendRepo domain.TrendingRepository,
	jwtCfg config.JWTConfig,
	logger *zap.Logger,
) *Service {
	return &Service{
		pg:          pg,
		bookRepo:    bookRepo,
		recRepo:     recRepo,
		sessionRepo: sessionRepo,
		cartRepo:    cartRepo,
		trendRepo:   trendRepo,
		jwtCfg:      jwtCfg,
		logger:      logger,
	}
}
