package tracing

import (
	"context"

	"github.com/opentracing/opentracing-go"
)

// StartRedisSpan redis开启追踪
func StartRedisSpan(ctx context.Context, cmd string) opentracing.Span {
	span, _ := opentracing.StartSpanFromContext(ctx, cmd)
	span.SetTag(DBType, "redis")
	span.SetTag(TraceID, GetTraceID(ctx))
	return span
}
