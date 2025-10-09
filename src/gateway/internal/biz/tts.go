package biz

import (
	"bufio"
	"context"
	"io"
	"strings"

	"github.com/Fl0rencess720/Doria/src/gateway/internal/pkgs/circuitbreaker"
	"github.com/Fl0rencess720/Doria/src/gateway/internal/pkgs/utils"
	ttsapi "github.com/Fl0rencess720/Doria/src/rpc/tts"
	"github.com/pion/webrtc/v3"
	"github.com/pion/webrtc/v3/pkg/media"
	"go.uber.org/zap"
)

type TTSRepo interface {
	CreateOfferPeerTrack(ctx context.Context, peerID, targetPeerID string) (*webrtc.TrackLocalStaticSample, error)
}

type ttsUseCase struct {
	repo           TTSRepo
	ttsClient      ttsapi.TTSServiceClient
	circuitBreaker *circuitbreaker.CircuitBreakerManager
}

func NewTTSUsecase(repo TTSRepo, ttsClient ttsapi.TTSServiceClient, cbManager *circuitbreaker.CircuitBreakerManager) TTSUseCase {
	return &ttsUseCase{
		repo:           repo,
		ttsClient:      ttsClient,
		circuitBreaker: cbManager,
	}
}

func (u *ttsUseCase) SynthesizeSpeech(ctx context.Context, reader io.Reader, sessionID string) error {
	_, err := u.circuitBreaker.Do(ctx, "tts-service.SynthesizeSpeech",
		func(ctx context.Context) (any, error) {
			track, err := u.repo.CreateOfferPeerTrack(ctx, utils.GenerateOfferPeerID(sessionID), utils.GenerateAnswerPeerID(sessionID))
			if err != nil {
				zap.L().Error("create offer peer track failed", zap.Error(err))
				return nil, err
			}

			scanner := bufio.NewScanner(reader)
			scanner.Split(utils.ScanOnPunctuation)

			for scanner.Scan() {
				text := strings.TrimSpace(scanner.Text())

				if text == "" {
					continue
				}

				audioContent, err := u.ttsClient.SynthesizeSpeech(ctx, &ttsapi.SynthesizeSpeechRequest{
					Text: text,
				})
				if err != nil {
					zap.L().Error("tts client error", zap.Error(err))
					return nil, err
				}

				if err := track.WriteSample(media.Sample{Data: audioContent.AudioContent}); err != nil {
					zap.L().Error("write sample error", zap.Error(err))
					return nil, err
				}
			}

			if err := scanner.Err(); err != nil {
				zap.L().Error("scanner error", zap.Error(err))
				return nil, err
			}
			return true, nil
		},
		func(ctx context.Context, err error) (any, error) {
			zap.L().Error("tts fallback triggered", zap.Error(err))
			return true, nil
		},
	)

	if err != nil {
		zap.L().Error("tts client error", zap.Error(err))
		return err
	}

	return nil
}
