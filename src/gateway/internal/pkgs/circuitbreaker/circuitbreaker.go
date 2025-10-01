package circuitbreaker

import "github.com/google/wire"

var ProviderSet = wire.NewSet(NewCircuitBreakerManager)
