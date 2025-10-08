package biz

import (
	"context"
	"encoding/json"

	"github.com/Fl0rencess720/Doria/src/gateway/internal/models"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

type SignalingRepo interface {
	RegisterAnswerPeer(peerID string, conn *websocket.Conn) error
	UnregisterAnswerPeer(peerID string) error
	RegisterOfferPeer(peerID string, conn *websocket.Conn) error
	UnregisterOfferPeer(peerID string) error
	GetAnswerPeer(peerID string) (*models.Peer, bool)
	GetOfferPeer(peerID string) (*models.Peer, bool)
}

type signalingUseCase struct {
	repo SignalingRepo
}

func NewSignalingUsecase(repo SignalingRepo) SignalingUseCase {
	return &signalingUseCase{
		repo: repo,
	}
}

func (u *signalingUseCase) RegisterAnswerPeer(ctx context.Context, conn *websocket.Conn, req *models.Request) error {
	peerID := req.SourceID

	if err := u.repo.RegisterAnswerPeer(peerID, conn); err != nil {
		zap.L().Error("failed to register answer peer", zap.String("peer_id", peerID), zap.Error(err))
		return err
	}

	resp := models.Response{Code: 0, Msg: "ok"}
	respBytes, err := json.Marshal(resp)
	if err != nil {
		zap.L().Error("failed to marshal response", zap.Error(err))
		return err
	}

	initRespMsg := models.Message{Cmd: models.CmdInitResp, Payload: respBytes}
	initRespBytes, err := json.Marshal(initRespMsg)
	if err != nil {
		zap.L().Error("failed to marshal init response message", zap.Error(err))
		return err
	}

	if err := conn.WriteMessage(websocket.TextMessage, initRespBytes); err != nil {
		zap.L().Error("websocket write error", zap.Error(err))
		return err
	}

	return nil
}

func (u *signalingUseCase) HandleAnswerPeerMessages(ctx context.Context, conn *websocket.Conn, sourcePeerID string) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		_, message, err := conn.ReadMessage()
		if err != nil {
			zap.L().Error("websocket read error", zap.Error(err))
			return err
		}

		var msg models.Message
		if err := json.Unmarshal(message, &msg); err != nil {
			zap.L().Error("failed to unmarshal message", zap.Error(err))
			continue
		}

		var req models.Request
		if err := json.Unmarshal(msg.Payload, &req); err != nil {
			zap.L().Error("failed to unmarshal request", zap.Error(err))
			continue
		}

		targetPeerID := req.TargetID
		targetPeer, exists := u.repo.GetOfferPeer(targetPeerID)
		if !exists {
			zap.L().Error("target offer peer not found", zap.String("target_peer_id", targetPeerID))
			continue
		}

		if err := targetPeer.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
			zap.L().Error("websocket write error", zap.Error(err))
			continue
		}

		resp := models.Response{Code: 0, Msg: "ok"}
		respBytes, _ := json.Marshal(resp)
		respMsg := models.Message{Cmd: msg.Cmd + 100, Payload: respBytes}
		respMsgBytes, _ := json.Marshal(respMsg)
		conn.WriteMessage(websocket.TextMessage, respMsgBytes)
	}
}

func (u *signalingUseCase) HandleOfferPeerMessages(ctx context.Context, conn *websocket.Conn, sourcePeerID string) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		_, message, err := conn.ReadMessage()
		if err != nil {
			zap.L().Error("read from offer peer error", zap.String("peer_id", sourcePeerID), zap.Error(err))
			return err
		}

		var msg models.Message
		if err := json.Unmarshal(message, &msg); err != nil {
			zap.L().Error("unmarshal message error", zap.Error(err))
			continue
		}

		var req models.Request
		if err := json.Unmarshal(msg.Payload, &req); err != nil {
			zap.L().Error("unmarshal request error", zap.Error(err))
			continue
		}

		if sourcePeerID == "" {
			sourcePeerID = req.SourceID
			zap.L().Info("offer peer registered", zap.String("peer_id", sourcePeerID))
		}

		if err := u.repo.RegisterOfferPeer(sourcePeerID, conn); err != nil {
			zap.L().Error("failed to register offer peer", zap.String("peer_id", sourcePeerID), zap.Error(err))
			continue
		}

		targetPeerID := req.TargetID
		targetPeer, exists := u.repo.GetAnswerPeer(targetPeerID)
		if !exists {
			zap.L().Error("target answer peer not found", zap.String("target_peer_id", targetPeerID))
			resp := models.Response{Code: 404, Msg: "target peer not found"}
			respBytes, _ := json.Marshal(resp)
			errorRespMsg := models.Message{Cmd: msg.Cmd + 100, Payload: respBytes}
			errorRespBytes, _ := json.Marshal(errorRespMsg)
			conn.WriteMessage(websocket.TextMessage, errorRespBytes)
			continue
		}

		zap.L().Info("forwarding message",
			zap.Int("cmd", msg.Cmd),
			zap.String("from", sourcePeerID),
			zap.String("to", targetPeerID))

		if err := targetPeer.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
			zap.L().Error("forward to answer peer error", zap.String("target_peer_id", targetPeerID), zap.Error(err))
			continue
		}

		resp := models.Response{Code: 0, Msg: "ok"}
		respBytes, _ := json.Marshal(resp)
		respMsg := models.Message{Cmd: msg.Cmd + 100, Payload: respBytes}
		respMsgBytes, _ := json.Marshal(respMsg)
		conn.WriteMessage(websocket.TextMessage, respMsgBytes)
	}
}
