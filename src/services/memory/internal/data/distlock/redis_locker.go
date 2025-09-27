package distlock

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type redisLocker struct {
	client *redis.Client
}

func NewRedisLocker(client *redis.Client) Locker {
	return &redisLocker{client: client}
}
func (r *redisLocker) Lock(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	return r.client.SetNX(ctx, key, "locked", ttl).Result()
}
func (r *redisLocker) Unlock(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}
