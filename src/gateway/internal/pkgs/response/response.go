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
	AuthError:   401,
}

var Message = map[uint]string{
	ServerError:       "服务端错误",
	FormError:         "参数错误",
	AuthError:         "认证失败",
	TokenExpired:      "Token已过期",
	LoginError:        "登录失败",
	RefreshTokenError: "刷新Token失败",
	RegisterError:     "注册失败",
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

func AuthErrorResponse(c *gin.Context, code uint, data ...any) {
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
