package tracing

import (
	"context"

	"github.com/opentracing/opentracing-go"
	"github.com/rs/xid"

	"lego-lib/net/http"
)

// StartHTTPClientSpan 根据ctx和http.Request获取span
func StartHTTPClientSpan(req *http.Request) opentracing.Span {
	traceid := GetTraceID(req.Context())
	req.Header.Set(TraceID, traceid)

	span, _ := opentracing.StartSpanFromContext(req.Context(), "http client")
	span.SetTag("http.url", req.URL.Host)
	span.SetTag("http.method", req.Method)
	span.SetTag(TraceID, traceid)

	span.Tracer().Inject(span.Context(), opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header))
	return span
}

// StartHTTPServerSpan 根据ctx和http.Request获取span
func StartHTTPServerSpan(req *http.Request) (opentracing.Span, *http.Request) {
	carrier := opentracing.HTTPHeadersCarrier(req.Header)

	spanctx, _ := GlobalTracer().Extract(opentracing.HTTPHeaders, carrier)
	span := opentracing.StartSpan(req.URL.Path, opentracing.ChildOf(spanctx))

	traceid := req.Header.Get(TraceID)
	if len(traceid) == 0 {
		traceid = xid.New().String()
	}
	span.SetTag(TraceID, traceid)

	c := context.WithValue(req.Context(), CtxTraceKey(TraceID), traceid)
	x := context.WithValue(c, CtxTraceKey("X-CLIENT-VERSION"), req.Header.Get("X-CLIENT-VERSION"))

	return span, req.WithContext(opentracing.ContextWithSpan(x, span))
}

// NewTraceMiddlewares http client tracing中间件
func NewTraceMiddlewares(packs ...string) http.Handler {
	return func(middle http.Middleware) {
		request := middle.GetRequest()
		TransPacks(request, packs...)

		span := StartHTTPClientSpan(request)
		middle.Next()
		span.Finish()
	}
}

// TransPacks ctx中指定数据（packs）写入到http header
func TransPacks(req *http.Request, packs ...string) {
	for _, pack := range packs {
		if s := GetTransPack(req.Context(), pack); s != "" {
			req.Header.Set(pack, s)
		}
	}
}
