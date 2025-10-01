package fallback

func FallbackStrategyProvider() FallbackStrategy {
	return NewDefaultDataFallback()
}
