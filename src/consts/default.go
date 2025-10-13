package consts

import "time"

var (
	DefaultLogFilePath = "./zap.log"

	ContentTypeMP3 = "audio/mp3"

	DoriaMemorySignalTopic = "doria_memory_signal"

	RedisSTMLengthKey = "stm_length"
	RedisMTMLengthKey = "mtm_length"
)

const (
	STMPageCachePrefix = "stm_pages"
	STMPageCacheTTL    = 12 * time.Hour
)
