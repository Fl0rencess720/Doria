package fallback

import "github.com/gin-gonic/gin"

type FallbackStrategy interface {
	Execute(c *gin.Context, serviceName string, err error)
}
