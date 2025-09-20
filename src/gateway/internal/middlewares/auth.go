package middlewares

import (
	"strings"

	"github.com/Fl0rencess720/Doria/src/gateway/internal/pkgs/jwtc"
	"github.com/Fl0rencess720/Doria/src/gateway/internal/pkgs/response"
	"github.com/gin-gonic/gin"
)

type ContextKey string

var (
	UserIDKey = ContextKey("user_id")
)

func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			response.AuthErrorResponse(c, response.AuthError)
			return
		}

		parts := strings.Split(tokenString, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.AuthErrorResponse(c, response.AuthError)
			return
		}

		parsedToken, isExpire, err := jwtc.ParseToken(parts[1])
		if err != nil {
			response.AuthErrorResponse(c, response.AuthError)
			return
		}
		if isExpire {
			response.AuthErrorResponse(c, response.TokenExpired)
			return
		}

		c.Set(string(UserIDKey), parsedToken.UserID)
		c.Next()
	}
}
