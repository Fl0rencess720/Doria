package response

import (
	"github.com/gin-gonic/gin"
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

	NoError
)

var HttpCode = map[ErrorCode]int{
	FormError:   400,
	ServerError: 500,
	AuthError:   401,
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
