package service

import (
	"context"

	ttsapi "github.com/Fl0rencess720/Doria/src/rpc/tts"
)

func (s *TTSService) SynthesizeSpeech(ctx context.Context, req *ttsapi.SynthesizeSpeechRequest) (*ttsapi.SynthesizeSpeechResponse, error) {
	audioContent, err := s.ttsUseCase.SynthesizeSpeech(req.Text)
	if err != nil {
		return nil, err
	}

	return &ttsapi.SynthesizeSpeechResponse{
		AudioContent: audioContent,
	}, nil
}
