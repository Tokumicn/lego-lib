package tracing

import (
	"context"

	"github.com/opentracing/opentracing-go"
)

// StartMysqlSpan mysql开启追踪
func StartMysqlSpan(ctx context.Context, cmd string) opentracing.Span {
	span, _ := opentracing.StartSpanFromContext(ctx, cmd)
	span.SetTag(DBType, "mysql")
	span.SetTag(TraceID, GetTraceID(ctx))
	return span
}
