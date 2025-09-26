package distlock

import (
	"context"
	"time"

	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(NewRedisLocker)

type Locker interface {
	Lock(ctx context.Context, key string, ttl time.Duration) (bool, error)
	Unlock(ctx context.Context, key string) error
}
