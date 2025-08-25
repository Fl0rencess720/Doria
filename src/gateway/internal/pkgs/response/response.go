package response

import (
	"github.com/gin-gonic/gin"
)

const (
	ServerError = iota
	FormError
	AuthError
	TokenExpired
	LoginError
	RefreshTokenError
	RegisterError
)

var HttpCode = map[uint]int{
	FormError:   400,
	ServerError: 500,
}

var Message = map[uint]string{
	ServerError: "服务端错误",
}

func SuccessResponse(c *gin.Context, data any) {
	c.JSON(200, gin.H{
		"msg":  "success",
		"code": 200,
		"data": data,
	})
}

func ErrorResponse(c *gin.Context, code uint, data ...any) {
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
