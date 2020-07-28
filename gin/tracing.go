package gin

import (
	"github.com/gin-gonic/gin"

	"github.com/Tokumicn/go-frame/lego-lib/tracing"
)

// Trace gin 分布式追踪插件
func Trace() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		span, request := tracing.StartHTTPServerSpan(ctx.Request)
		ctx.Request = request

		ctx.Next()

		span.SetTag(tracing.HTTPStatusCode, uint16(ctx.Writer.Status()))
		span.Finish()
	}
}
