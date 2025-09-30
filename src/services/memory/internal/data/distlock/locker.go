package distlock

import (
	"context"
	"errors"
	"time"
)

var ErrLockNotAcquired = errors.New("lock: not acquired by another process")

type Locker interface {
	Lock(ctx context.Context, key string, ttl time.Duration) error
	Unlock(ctx context.Context, key string) error
}
