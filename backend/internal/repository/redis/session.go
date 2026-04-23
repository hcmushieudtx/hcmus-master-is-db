package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	sessionPrefix   = "session:"   // session:<userID> → token string
	blacklistPrefix = "blacklist:" // blacklist:<token> → "1"
	defaultTTL      = 24 * time.Hour
)

// SessionRepository implements domain.SessionRepository against Redis.
type SessionRepository struct {
	rdb *client
}

// NewSessionRepository creates a SessionRepository.
func NewSessionRepository(rdb *redis.Client) *SessionRepository {
	return &SessionRepository{rdb: rdb}
}

// SetToken stores the JWT token for a user with a TTL.
func (r *SessionRepository) SetToken(ctx context.Context, userID string, token string) error {
	key := sessionPrefix + userID
	return r.rdb.Set(ctx, key, token, defaultTTL).Err()
}

// GetToken retrieves the active token for a user.
func (r *SessionRepository) GetToken(ctx context.Context, userID string) (string, error) {
	key := sessionPrefix + userID
	val, err := r.rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil
	}
	return val, err
}

// BlacklistToken adds a token to the blacklist so that further requests using
// it are rejected even before its JWT expiry time.
func (r *SessionRepository) BlacklistToken(ctx context.Context, token string) error {
	key := fmt.Sprintf("%s%s", blacklistPrefix, token)
	return r.rdb.Set(ctx, key, "1", defaultTTL).Err()
}

// IsBlacklisted returns true if the given token has been revoked.
func (r *SessionRepository) IsBlacklisted(ctx context.Context, token string) (bool, error) {
	key := fmt.Sprintf("%s%s", blacklistPrefix, token)
	val, err := r.rdb.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return val > 0, nil
}
