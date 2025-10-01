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

func (m *CircuitBreakerManager) GetBreaker(serviceName string) *gobreaker.CircuitBreaker {
	if breaker, exists := m.breakers[serviceName]; exists {
		return breaker
	}

	settings := gobreaker.Settings{
		Name:        serviceName,
		MaxRequests: 3,
		Interval:    time.Second * 60,
		Timeout:     time.Second * 60,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= 3 && failureRatio >= 0.6
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			zap.L().Info("CircuitBreaker state changed",
				zap.String("service", name),
				zap.String("from", from.String()),
				zap.String("to", to.String()),
			)
		},
	}

	breaker := gobreaker.NewCircuitBreaker(settings)
	m.breakers[serviceName] = breaker
	return breaker
}

func (m *CircuitBreakerManager) CallWithBreaker(serviceName string, fn func() (any, error)) (any, error) {
	breaker := m.GetBreaker(serviceName)
	return breaker.Execute(fn)
}

func (m *CircuitBreakerManager) CallWithBreakerContext(ctx context.Context, serviceName string, fn func() (any, error)) (any, error) {
	breaker := m.GetBreaker(serviceName)

	wrappedFn := func() (any, error) {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		return fn()
	}

	return breaker.Execute(wrappedFn)
}
