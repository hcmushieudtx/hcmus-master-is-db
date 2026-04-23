package redis

import "github.com/redis/go-redis/v9"

// client is a thin alias kept here so sub-packages share the same import.
type client = redis.Client
