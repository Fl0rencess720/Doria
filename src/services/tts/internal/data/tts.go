package data

import (
	"github.com/Fl0rencess720/Doria/src/services/tts/internal/biz"
)

type ttsRepo struct {
}

func NewTTSRepo() biz.TTSRepo {
	return &ttsRepo{}
}
