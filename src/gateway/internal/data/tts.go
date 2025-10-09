package data

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/Fl0rencess720/Doria/src/common/registry"
	"github.com/Fl0rencess720/Doria/src/gateway/internal/biz"
	"github.com/Fl0rencess720/Doria/src/gateway/internal/models"
	ttsapi "github.com/Fl0rencess720/Doria/src/rpc/tts"
	"github.com/gorilla/websocket"
	_ "github.com/mbobakov/grpc-consul-resolver"
	"github.com/pion/webrtc/v3"
	"github.com/spf13/viper"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type SignalingManager struct {
	WSConn      *websocket.Conn
	PeerConn    *webrtc.PeerConnection
	TrackReady  chan struct{}
	ErrorChan   chan error
	DoneChan    chan struct{}
	IsConnected bool
	mutex       sync.RWMutex
}

type ConnectionState struct {
	State       string
	LastUpdated int64
	ErrorCount  int
}

type TTSRepo struct {
	signalingURL string
}

func NewTTSRepo() biz.TTSRepo {
	return &TTSRepo{
		signalingURL: viper.GetString("webrtc.signaling_offer_url"),
	}
}

func NewTTSClient() ttsapi.TTSServiceClient {
	discoveryManager := registry.NewDiscoveryManager()

	conn, err := discoveryManager.CreateGrpcConnection(
		context.Background(),
		"doria-tts",
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		zap.L().Panic("new grpc client failed", zap.Error(err))
	}

	client := ttsapi.NewTTSServiceClient(conn)
	return client
}

func (r *TTSRepo) CreateOfferPeerTrack(ctx context.Context, peerID, targetPeerID string) (*webrtc.TrackLocalStaticSample, error) {
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	}

	peerConnection, err := webrtc.NewPeerConnection(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create peer connection: %w", err)
	}

	track, err := webrtc.NewTrackLocalStaticSample(
		webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeOpus},
		"audio",
		"pion")
	if err != nil {
		peerConnection.Close()
		return nil, fmt.Errorf("failed to create track: %w", err)
	}

	if _, err = peerConnection.AddTrack(track); err != nil {
		peerConnection.Close()
		return nil, fmt.Errorf("failed to add track: %w", err)
	}

	wsConn, _, err := websocket.DefaultDialer.Dial(r.signalingURL, nil)
	if err != nil {
		peerConnection.Close()
		return nil, fmt.Errorf("failed to connect to signaling server: %w", err)
	}

	signalingManager := NewSignalingManager()
	signalingManager.PeerConn = peerConnection
	signalingManager.WSConn = wsConn

	peerConnection.OnICECandidate(func(c *webrtc.ICECandidate) {
		if c == nil {
			zap.L().Info("offer: ICE candidate is nil")
			return
		}

		candidateBytes, err := json.Marshal(c.ToJSON())
		if err != nil {
			zap.L().Error("offer: marshal candidate error", zap.Error(err))
			return
		}

		if err := r.offerMessage(wsConn, models.CmdCandidate, &models.Request{
			SourceID: peerID,
			TargetID: targetPeerID,
			Body:     candidateBytes,
		}); err != nil {
			zap.L().Error("offer: send candidate failed", zap.Error(err))
			return
		}
	})

	peerConnection.OnConnectionStateChange(func(s webrtc.PeerConnectionState) {
		zap.L().Info("offer: Peer Connection State has changed", zap.String("state", s.String()))

		switch s {
		case webrtc.PeerConnectionStateConnected:
			signalingManager.SetConnected(true)
		case webrtc.PeerConnectionStateFailed, webrtc.PeerConnectionStateClosed, webrtc.PeerConnectionStateDisconnected:
			signalingManager.SetConnected(false)
		}
	})

	offer, err := peerConnection.CreateOffer(nil)
	if err != nil {
		peerConnection.Close()
		wsConn.Close()
		return nil, fmt.Errorf("failed to create offer: %w", err)
	}

	if err := peerConnection.SetLocalDescription(offer); err != nil {
		peerConnection.Close()
		wsConn.Close()
		return nil, fmt.Errorf("failed to set local description: %w", err)
	}

	offerBytes, err := json.Marshal(offer)
	if err != nil {
		peerConnection.Close()
		wsConn.Close()
		return nil, fmt.Errorf("failed to marshal offer: %w", err)
	}

	if err := r.offerMessage(wsConn, models.CmdOffer, &models.Request{
		SourceID: peerID,
		TargetID: targetPeerID,
		Body:     offerBytes,
	}); err != nil {
		peerConnection.Close()
		wsConn.Close()
		return nil, fmt.Errorf("failed to send offer: %w", err)
	}

	if err := signalingManager.Start(); err != nil {
		peerConnection.Close()
		wsConn.Close()
		return nil, fmt.Errorf("failed to start signaling manager: %w", err)
	}

	if err := signalingManager.WaitForConnection(30 * time.Second); err != nil {
		peerConnection.Close()
		wsConn.Close()
		signalingManager.Cleanup()
		return nil, fmt.Errorf("connection establishment failed: %w", err)
	}

	zap.L().Info("offer: WebRTC connection established successfully")

	go func() {
		<-ctx.Done()
		zap.L().Info("offer: cleaning up resources due to context cancellation")
		signalingManager.Cleanup()
	}()

	return track, nil
}

func (r *TTSRepo) offerMessage(conn *websocket.Conn, cmd int, req *models.Request) error {
	reqBytes, _ := json.Marshal(req)
	msg := models.Message{Cmd: cmd, Payload: reqBytes}
	msgBytes, _ := json.Marshal(msg)

	if err := conn.WriteMessage(websocket.TextMessage, msgBytes); err != nil {
		return err
	}

	return nil
}

func NewSignalingManager() *SignalingManager {
	return &SignalingManager{
		TrackReady:  make(chan struct{}, 1),
		ErrorChan:   make(chan error, 1),
		DoneChan:    make(chan struct{}),
		IsConnected: false,
	}
}

func (sm *SignalingManager) Start() error {
	go sm.handleSignalingMessages()
	return nil
}

func (sm *SignalingManager) handleSignalingMessages() {
	defer func() {
		if r := recover(); r != nil {
			zap.L().Error("SignalingManager: panic in message handler", zap.Any("panic", r))
			sm.ErrorChan <- fmt.Errorf("panic in message handler: %v", r)
		}
	}()

	for {
		select {
		case <-sm.DoneChan:
			zap.L().Info("SignalingManager: stopping message handler")
			return
		default:
			if sm.WSConn == nil {
				err := fmt.Errorf("WebSocket connection is nil")
				zap.L().Error("SignalingManager: WebSocket connection is nil")
				sm.ErrorChan <- err
				return
			}

			if err := sm.WSConn.SetReadDeadline(time.Now().Add(60 * time.Second)); err != nil {
				zap.L().Error("SignalingManager: failed to set read deadline", zap.Error(err))
				sm.ErrorChan <- err
				return
			}

			_, message, err := sm.WSConn.ReadMessage()
			if err != nil {
				zap.L().Error("SignalingManager: read message error", zap.Error(err))
				sm.ErrorChan <- err
				return
			}

			var msg models.Message
			if err := json.Unmarshal(message, &msg); err != nil {
				zap.L().Error("SignalingManager: unmarshal message error", zap.Error(err), zap.ByteString("message", message))
				continue
			}

			if err := sm.processMessage(&msg); err != nil {
				zap.L().Error("SignalingManager: process message error", zap.Error(err), zap.Int("cmd", msg.Cmd))
				sm.ErrorChan <- err
			}
		}
	}
}

func (sm *SignalingManager) processMessage(msg *models.Message) error {
	switch msg.Cmd {
	case models.CmdAnswer:
		return sm.handleAnswer(msg)
	case models.CmdCandidate:
		return sm.handleCandidate(msg)
	case models.CmdAnswerResp, models.CmdCandidateResp, models.CmdOfferResp:
		return nil
	default:
		zap.L().Info("SignalingManager: unknown message", zap.Int("cmd", msg.Cmd))
		return nil
	}
}

func (sm *SignalingManager) handleAnswer(msg *models.Message) error {
	var req models.Request
	if err := json.Unmarshal(msg.Payload, &req); err != nil {
		return fmt.Errorf("unmarshal request error: %w", err)
	}

	answer := webrtc.SessionDescription{}
	if err := json.Unmarshal(req.Body, &answer); err != nil {
		return fmt.Errorf("unmarshal answer error: %w", err)
	}

	if err := sm.PeerConn.SetRemoteDescription(answer); err != nil {
		return fmt.Errorf("set remote description error: %w", err)
	}

	select {
	case sm.TrackReady <- struct{}{}:
	default:
	}
	return nil
}

func (sm *SignalingManager) handleCandidate(msg *models.Message) error {
	var req models.Request
	if err := json.Unmarshal(msg.Payload, &req); err != nil {
		return fmt.Errorf("unmarshal request error: %w", err)
	}

	var candidate webrtc.ICECandidateInit
	if err := json.Unmarshal(req.Body, &candidate); err != nil {
		return fmt.Errorf("unmarshal candidate error: %w", err)
	}

	if err := sm.PeerConn.AddICECandidate(candidate); err != nil {
		return fmt.Errorf("add ICE candidate error: %w", err)
	}

	return nil
}

func (sm *SignalingManager) SetConnected(connected bool) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	sm.IsConnected = connected
}

func (sm *SignalingManager) IsConnectedStatus() bool {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	return sm.IsConnected
}

func (sm *SignalingManager) Cleanup() {
	close(sm.DoneChan)

	if sm.WSConn != nil {
		sm.WSConn.Close()
	}

	if sm.PeerConn != nil {
		sm.PeerConn.Close()
	}
}

func (sm *SignalingManager) WaitForConnection(timeout time.Duration) error {
	select {
	case <-sm.TrackReady:
		return nil
	case err := <-sm.ErrorChan:
		zap.L().Error("SignalingManager: connection failed", zap.Error(err))
		return err
	case <-time.After(timeout):
		return fmt.Errorf("connection timeout after %v", timeout)
	}
}
