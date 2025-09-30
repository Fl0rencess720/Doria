package distlock

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const unlockScript = `
if redis.call("GET", KEYS[1]) == ARGV[1] then
    return redis.call("DEL", KEYS[1])
else
    return 0
end
`
const refreshScript = `
if redis.call("GET", KEYS[1]) == ARGV[1] then
    return redis.call("PEXPIRE", KEYS[1], ARGV[2])
else
    return 0
end
`

type activeLock struct {
	value    string
	stopChan chan struct{}
}

type redisLocker struct {
	client      *redis.Client
	mu          sync.Mutex
	activeLocks map[string]*activeLock
}

func NewRedisLocker(client *redis.Client) Locker {
	return &redisLocker{
		client:      client,
		activeLocks: make(map[string]*activeLock),
	}
}

func (r *redisLocker) Lock(ctx context.Context, key string, ttl time.Duration) error {
	r.mu.Lock()
	if _, ok := r.activeLocks[key]; ok {
		r.mu.Unlock()
		return errors.New("lock: already locked by this instance")
	}
	r.mu.Unlock()

	value := uuid.NewString()

	ok, err := r.client.SetNX(ctx, key, value, ttl).Result()
	if err != nil {
		return err
	}
	if !ok {
		return ErrLockNotAcquired
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.activeLocks[key]; ok {
		r.client.Eval(ctx, unlockScript, []string{key}, value)
		return errors.New("lock: already locked by this instance")
	}

	lockState := &activeLock{
		value:    value,
		stopChan: make(chan struct{}),
	}
	r.activeLocks[key] = lockState

	go r.watchdog(key, value, ttl, lockState.stopChan)

	return nil
}

func (r *redisLocker) Unlock(ctx context.Context, key string) error {
	r.mu.Lock()
	lockState, ok := r.activeLocks[key]
	if !ok {
		r.mu.Unlock()
		return errors.New("lock: not locked by this instance")
	}

	delete(r.activeLocks, key)
	close(lockState.stopChan)
	r.mu.Unlock()

	_, err := r.client.Eval(ctx, unlockScript, []string{key}, lockState.value).Result()
	if err != nil && err != redis.Nil {
		return fmt.Errorf("failed to unlock key '%s': %w", key, err)
	}

	return nil
}

func (r *redisLocker) watchdog(key, value string, ttl time.Duration, stopChan <-chan struct{}) {
	ticker := time.NewTicker(ttl / 3)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			res, err := r.client.Eval(ctx, refreshScript, []string{key}, value, ttl.Milliseconds()).Int64()
			cancel()

			if err != nil || res == 0 {
				return
			}
		case <-stopChan:
			return
		}
	}
}
