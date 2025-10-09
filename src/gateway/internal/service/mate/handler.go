package mate

import (
	"io"
	"net/http"
	"strconv"

	"github.com/Fl0rencess720/Doria/src/gateway/internal/biz"
	"github.com/Fl0rencess720/Doria/src/gateway/internal/models"
	"github.com/Fl0rencess720/Doria/src/gateway/internal/pkgs/response"
	"github.com/Fl0rencess720/Doria/src/gateway/internal/service/middlewares"

	"github.com/gin-contrib/sse"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type MateHandler struct {
	mateUseCase biz.MateUseCase
	ttsUseCase  biz.TTSUseCase
}

func NewMateHandler(mateUseCase biz.MateUseCase, ttsUseCase biz.TTSUseCase) *MateHandler {
	return &MateHandler{
		mateUseCase: mateUseCase,
		ttsUseCase:  ttsUseCase,
	}
}

func (u *MateHandler) Chat(c *gin.Context) {
	ctx := c.Request.Context()

	userID := c.GetInt(string(middlewares.UserIDKey))

	req := &models.ChatReq{}
	if err := c.ShouldBindJSON(&req); err != nil {
		zap.L().Warn("request bind error", zap.Error(err))
		response.ErrorResponse(c, response.FormError)
		return
	}

	output, errorCode, err := u.mateUseCase.Chat(ctx, req, userID)
	if err != nil {
		zap.L().Error("chat error", zap.Error(err))
		response.ErrorResponse(c, errorCode)
		return
	}

	response.SuccessResponse(c, output)
}

func (u *MateHandler) ChatStream(c *gin.Context) {
	ctx := c.Request.Context()

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		zap.L().Error("ResponseWriter does not support Flusher")
		return
	}

	userID := c.GetInt(string(middlewares.UserIDKey))
	req := &models.ChatReq{}

	if err := c.ShouldBindJSON(&req); err != nil {
		zap.L().Warn("request bind error", zap.Error(err))
		response.SendSSEError(c.Writer, flusher, "FormError", err.Error())
		return
	}

	stream, err := u.mateUseCase.CreateChatStream(ctx, req, userID)
	if err != nil {
		zap.L().Error("create chat stream error", zap.Error(err))
		response.SendSSEError(c.Writer, flusher, "ServerError", err.Error())
		return
	}

	pr, pw := io.Pipe()
	defer pw.Close()

	go func() {
		u.ttsUseCase.SynthesizeSpeech(ctx, pr)
	}()

	for {
		select {
		case <-ctx.Done():
			zap.L().Info("Client connection closed", zap.Error(ctx.Err()))
			return
		default:
		}

		resp, err := stream.Recv()
		if err == io.EOF {
			zap.L().Info("gRPC stream finished successfully")
			return
		}

		if err != nil {
			zap.L().Error("failed to receive from gRPC stream", zap.Error(err))
			response.SendSSEError(c.Writer, flusher, "gRPCError", err.Error())
			return
		}

		_, err = pw.Write([]byte(resp.Content))
		if err != nil {
			zap.L().Error("failed to write to pipe", zap.Error(err))
			response.SendSSEError(c.Writer, flusher, "ServerError", err.Error())
			return
		}

		data := map[string]interface{}{
			"type":       "chunk",
			"content":    resp.Content,
			"message_id": resp.MessageId,
			"timestamp":  resp.Timestamp,
			"finished":   resp.Finished,
		}

		if err = sse.Encode(c.Writer, sse.Event{
			Event: "message",
			Data:  data,
		}); err != nil {
			zap.L().Error("Error writing to SSE stream (client disconnected?)", zap.Error(err))
			return
		}

		flusher.Flush()

		if resp.Finished {
			zap.L().Info("Chat response finished")
			return
		}
	}
}

func (u *MateHandler) GetUserPages(c *gin.Context) {
	ctx := c.Request.Context()

	userID := c.GetInt(string(middlewares.UserIDKey))

	req := &models.GetUserPagesRequest{
		UserID:   userID,
		Cursor:   c.Query("cursor"),
		PageSize: 20,
	}

	if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
		if pageSize, err := strconv.Atoi(pageSizeStr); err == nil && pageSize > 0 {
			req.PageSize = pageSize
			if pageSize > 100 {
				req.PageSize = 100
			}
		}
	}

	pages, errorCode, err := u.mateUseCase.GetUserPages(ctx, req)
	if err != nil {
		zap.L().Error("GetUserPages error", zap.Error(err))
		response.ErrorResponse(c, errorCode)
		return
	}

	response.SuccessResponse(c, pages)
}
