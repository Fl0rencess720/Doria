package middlewares

import (
	"github.com/Fl0rencess720/Doria/src/common/tracing"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/attribute"
)

func Trace() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, span := tracing.Tracer.Start(c.Request.Context(), "HTTP "+c.Request.Method+" "+c.Request.URL.Path)
		defer span.End()

		span.SetAttributes(
			attribute.String("http.method", c.Request.Method),
			attribute.String("http.url", c.Request.URL.String()),
			attribute.String("http.host", c.Request.Host),
		)

		c.Request = c.Request.WithContext(ctx)
		c.Next()

		span.SetAttributes(attribute.Int("http.status_code", c.Writer.Status()))
	}
}
