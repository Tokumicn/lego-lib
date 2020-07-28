package tracing

import (
	"context"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
)

type CtxTraceKey string

const (
	// TraceID 系统内部统一的trace id 标识
	TraceID = "Ht-Trace-Id"
	// HTTPMethod http 方法标识
	HTTPMethod = "http.method"
	// HTTPUrl http url标识
	HTTPUrl = "http.url"
	// HTTPStatusCode http 状态码标识
	HTTPStatusCode = "http.status_code"
	// DBType 数据库类型标识
	DBType = "db.type"
)

var (
	// DefaultFlushInterval 数据上报 刷新周期 默认配置
	DefaultFlushInterval = 10 * time.Second
	// DefaultAgentAddress  数据上报 默认地址
	DefaultAgentAddress = "127.0.0.1:6831"

	// NilSamplerRate 追踪数据采样比率（SamplerTypeConst）  0%采样
	NilSamplerRate = 0.00
	// LowSamplerRate 追踪数据采样比率（SamplerTypeConst）  5%采样
	LowSamplerRate = 0.05
	// MidSamplerRate 追踪数据采样比率（SamplerTypeConst） 20%采样
	MidSamplerRate = 0.20
	// HigSamplerRate 追踪数据采样比率（SamplerTypeConst） 60%采样
	HigSamplerRate = 0.60
	// AllSamplerRate 追踪数据采样比率（SamplerTypeConst）100%采样
	AllSamplerRate = 1.00
)

// Init 分布式追踪初始化函数
func Init(service string, rate float64) error {
	c := config.Configuration{
		Sampler: &config.SamplerConfig{
			Type:  jaeger.SamplerTypeConst,
			Param: rate,
		},
		Reporter: &config.ReporterConfig{
			LogSpans:            false,
			BufferFlushInterval: DefaultFlushInterval,
			LocalAgentHostPort:  DefaultAgentAddress,
		},
	}

	_, err := c.InitGlobalTracer(service)
	return err
}

// GlobalTracer 获取opentracing全局对象
func GlobalTracer() opentracing.Tracer {
	return opentracing.GlobalTracer()
}

// GetTraceID 根据ctx获取 追踪id
func GetTraceID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	id, _ := ctx.Value(CtxTraceKey(TraceID)).(string)
	return id
}

// GetTransPack 根据ctx获取指定的数据
func GetTransPack(ctx context.Context, name string) string {
	if ctx == nil {
		return ""
	}

	val, _ := ctx.Value(CtxTraceKey(name)).(string)
	return val
}
