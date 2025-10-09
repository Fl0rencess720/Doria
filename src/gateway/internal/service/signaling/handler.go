package signaling

import (
	"encoding/json"
	"net/http"

	"github.com/Fl0rencess720/Doria/src/gateway/internal/biz"
	"github.com/Fl0rencess720/Doria/src/gateway/internal/models"
	"github.com/Fl0rencess720/Doria/src/gateway/internal/pkgs/response"
	"github.com/Fl0rencess720/Doria/src/gateway/internal/pkgs/utils"
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

	sessionID := utils.GenerateUniqueID()
	peerID := utils.GenerateAnswerPeerID(sessionID)

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

	if msg.Cmd == models.CmdInit {
		respMsg := models.Message{
			Cmd: models.CmdInitResp,
			Payload: func() []byte {
				resp, _ := json.Marshal(models.Response{
					Code: 0,
					Msg:  "ok",
				})
				return resp
			}(),
		}
		if err := conn.WriteJSON(respMsg); err != nil {
			zap.L().Error("failed to send init response", zap.Error(err))
			return
		}
	}

	req := &models.Request{
		SourceID: peerID,
	}

	if err := h.signalingUseCase.RegisterAnswerPeer(ctx, conn, req); err != nil {
		zap.L().Error("failed to register answer peer", zap.String("peer_id", peerID), zap.Error(err))
		response.SendWebSocketError(conn, response.ServerError)
		return
	}

	defer func() {
		zap.L().Info("answer peer unregistered", zap.String("peer_id", peerID))
	}()

	peerIDPayload := map[string]interface{}{
		"peer_id": peerID,
		"status":  "registered",
	}

	peerIDPayloadBytes, err := json.Marshal(peerIDPayload)
	if err != nil {
		zap.L().Error("failed to marshal peer_id payload", zap.Error(err))
		return
	}

	registerRespMsg := models.Message{
		Cmd:     models.CmdRegisterResp,
		Payload: peerIDPayloadBytes,
	}

	if err := conn.WriteJSON(registerRespMsg); err != nil {
		zap.L().Error("failed to send register response with peer_id", zap.Error(err))
		return
	}

	zap.L().Info("answer peer registered successfully", zap.String("peer_id", peerID))

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
