package signaling

import (
	"encoding/json"
	"net/http"

	"github.com/Fl0rencess720/Doria/src/gateway/internal/biz"
	"github.com/Fl0rencess720/Doria/src/gateway/internal/models"
	"github.com/Fl0rencess720/Doria/src/gateway/internal/pkgs/response"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type SignalingHandler struct {
	signalingUseCase biz.SignalingUseCase
}

func NewSignalingHandler(signalingUseCase biz.SignalingUseCase) *SignalingHandler {
	return &SignalingHandler{
		signalingUseCase: signalingUseCase,
	}
}

func (h *SignalingHandler) Register(c *gin.Context) {
	ctx := c.Request.Context()

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		zap.L().Error("websocket upgrade error", zap.Error(err))
		response.ErrorResponse(c, response.ServerError)
		return
	}
	defer conn.Close()

	_, message, err := conn.ReadMessage()
	if err != nil {
		zap.L().Error("websocket read error", zap.Error(err))
		response.ErrorResponse(c, response.ServerError)
		return
	}

	var msg models.Message
	if err := json.Unmarshal(message, &msg); err != nil {
		zap.L().Error("failed to unmarshal message", zap.Error(err))
		response.SendWebSocketError(conn, response.ServerError)
		return
	}

	if msg.Cmd != models.CmdInit {
		zap.L().Error("invalid command", zap.Int("cmd", msg.Cmd))
		response.SendWebSocketError(conn, response.ServerError)
		return
	}

	var req models.Request
	if err := json.Unmarshal(msg.Payload, &req); err != nil {
		zap.L().Error("failed to unmarshal request", zap.Error(err))
		response.SendWebSocketError(conn, response.ServerError)
		return
	}

	peerID := req.SourceID

	if err := h.signalingUseCase.RegisterAnswerPeer(ctx, conn, &req); err != nil {
		zap.L().Error("failed to register answer peer", zap.String("peer_id", peerID), zap.Error(err))
		response.SendWebSocketError(conn, response.ServerError)
		return
	}

	defer func() {
		zap.L().Info("answer peer unregistered", zap.String("peer_id", peerID))
	}()

	if err := h.signalingUseCase.HandleAnswerPeerMessages(ctx, conn, peerID); err != nil {
		zap.L().Error("error handling answer peer messages", zap.String("peer_id", peerID), zap.Error(err))
		response.SendWebSocketError(conn, response.ServerError)
	}
}

func (h *SignalingHandler) Offer(c *gin.Context) {
	ctx := c.Request.Context()

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		zap.L().Error("websocket upgrade error", zap.Error(err))
		response.ErrorResponse(c, response.ServerError)
		return
	}
	defer conn.Close()

	var sourcePeerID string

	defer func() {
		if sourcePeerID != "" {
			zap.L().Info("offer peer unregistered", zap.String("peer_id", sourcePeerID))
		}
	}()

	if err := h.signalingUseCase.HandleOfferPeerMessages(ctx, conn, sourcePeerID); err != nil {
		zap.L().Error("error handling offer peer messages", zap.String("peer_id", sourcePeerID), zap.Error(err))
		response.SendWebSocketError(conn, response.ServerError)
	}
}
