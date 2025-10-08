package response

import (
	"net/http"

	"github.com/gin-contrib/sse"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type ErrorCode uint

const (
	ServerError ErrorCode = iota
	FormError
	AuthError
	TokenExpired
	RefreshTokenError

	UserExistError
	CodeError

	UserNotExistError
	PasswordError

	RateLimitError
	DegradedError

	NoError
)

var HttpCode = map[ErrorCode]int{
	FormError:      400,
	ServerError:    500,
	AuthError:      401,
	RateLimitError: 429,
	DegradedError:  503,
}

var Message = map[ErrorCode]string{
	ServerError:       "服务端错误",
	FormError:         "参数错误",
	AuthError:         "认证失败",
	TokenExpired:      "Token已过期",
	RefreshTokenError: "刷新Token失败",

	UserExistError: "用户已存在",
	CodeError:      "验证码错误",

	UserNotExistError: "用户不存在",
	PasswordError:     "密码错误",
	RateLimitError:    "请求过于频繁",
	DegradedError:     "服务暂时不可用",
}

func SuccessResponse(c *gin.Context, data any) {
	c.JSON(200, gin.H{
		"msg":  "success",
		"code": 200,
		"data": data,
	})
}

func ErrorResponse(c *gin.Context, code ErrorCode, data ...any) {
	httpStatus, ok := HttpCode[code]
	if !ok {
		httpStatus = 403
	}
	msg, ok := Message[code]
	if !ok {
		msg = "未知错误"
	}

	c.JSON(httpStatus, gin.H{
		"code": code,
		"msg":  msg,
	})
}

func AuthErrorResponse(c *gin.Context, code ErrorCode, data ...any) {
	httpStatus, ok := HttpCode[code]
	if !ok {
		httpStatus = 401
	}
	msg, ok := Message[code]
	if !ok {
		msg = "未知错误"
	}
	c.AbortWithStatusJSON(httpStatus, gin.H{
		"code": code,
		"msg":  msg,
	})

}

func SendSSEError(w http.ResponseWriter, flusher http.Flusher, eventName string, errMsg string) {
	errData := map[string]interface{}{
		"code": http.StatusInternalServerError,
		"msg":  "Stream Error: " + errMsg,
	}

	if err := sse.Encode(w, sse.Event{
		Event: eventName,
		Data:  errData,
	}); err != nil {
		zap.L().Error("Error encoding final SSE error event", zap.Error(err))
	}

	flusher.Flush()
}
