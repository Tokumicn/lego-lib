package tracing

import (
	"context"

	"github.com/Shopify/sarama"
	"github.com/opentracing/opentracing-go"
	"github.com/rs/xid"
)

// StartKafkaProducerSpan kafka producer开启追踪
func StartKafkaProducerSpan(ctx context.Context, topic string) (opentracing.Span, []sarama.RecordHeader) {
	traceid := GetTraceID(ctx)
	if len(traceid) == 0 {
		traceid = xid.New().String()
	}

	span, _ := opentracing.StartSpanFromContext(ctx, "producer:"+topic)
	span.SetTag("broker", "kafka producer")
	span.SetTag(TraceID, traceid)

	carrier := opentracing.TextMapCarrier{}
	span.Tracer().Inject(span.Context(), opentracing.TextMap, carrier)

	headers := make([]sarama.RecordHeader, 0)
	for k, v := range carrier {
		headers = append(headers, sarama.RecordHeader{Key: []byte(k), Value: []byte(v)})
	}

	headers = append(headers, sarama.RecordHeader{Key: []byte(TraceID), Value: []byte(traceid)})
	return span, headers
}

// StartKafkaConsumerSpan kafka consumer开启追踪
func StartKafkaConsumerSpan(headers []*sarama.RecordHeader, topic string) (opentracing.Span, context.Context) {
	carrier := opentracing.TextMapCarrier{}
	for _, header := range headers {
		carrier[string(header.Key)] = string(header.Value)
	}
	traceid, _ := carrier[TraceID]

	spanctx, _ := GlobalTracer().Extract(opentracing.TextMap, carrier)
	span := opentracing.StartSpan("consumer:"+topic, opentracing.FollowsFrom(spanctx))
	span.SetTag("broker", "kafka consumer")
	span.SetTag(TraceID, traceid)

	ctx := opentracing.ContextWithSpan(context.Background(), span)
	return span, context.WithValue(ctx, CtxTraceKey(TraceID), traceid)
}
