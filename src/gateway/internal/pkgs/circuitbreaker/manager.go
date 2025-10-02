package circuitbreaker

import (
	"context"
	"time"

	"github.com/sony/gobreaker"
	"go.uber.org/zap"
)

type CircuitBreakerManager struct {
	breakers map[string]*gobreaker.CircuitBreaker
}

func NewCircuitBreakerManager() *CircuitBreakerManager {
	return &CircuitBreakerManager{
		breakers: make(map[string]*gobreaker.CircuitBreaker),
	}
}

func (m *CircuitBreakerManager) GetBreaker(key string) *gobreaker.CircuitBreaker {
	if breaker, exists := m.breakers[key]; exists {
		return breaker
	}

	settings := gobreaker.Settings{
		Name:        key,
		MaxRequests: 3,
		Interval:    time.Second * 60,
		Timeout:     time.Second * 60,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= 3 && failureRatio >= 0.6
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			zap.L().Info("CircuitBreaker state changed",
				zap.String("interface", name),
				zap.String("from", from.String()),
				zap.String("to", to.String()),
			)
		},
	}

	breaker := gobreaker.NewCircuitBreaker(settings)
	m.breakers[key] = breaker
	return breaker
}

func (m *CircuitBreakerManager) Do(
	ctx context.Context,
	key string,
	fn func(ctx context.Context) (any, error),
	fallback func(ctx context.Context, err error) (any, error),
) (any, error) {
	breaker := m.GetBreaker(key)

	wrappedFn := func() (any, error) {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		return fn(ctx)
	}

	result, err := breaker.Execute(wrappedFn)
	if err != nil {
		if fallback != nil {
			return fallback(ctx, err)
		}
		return nil, err
	}

	return result, nil
}
