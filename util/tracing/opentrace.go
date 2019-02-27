package tracing

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"

	"github.com/Axway/ace-golang-sdk/util/logging"
	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	opentracingLog "github.com/opentracing/opentracing-go/log"
	jaeger "github.com/uber/jaeger-client-go"
	jaegerconfig "github.com/uber/jaeger-client-go/config"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

const (
	// OpentracingContext const
	OpentracingContext string = "opentracingContext"
)

// set it via configuration
var opentracingEnabled = true

// TrLog - wrapper around file-based logger
var trLog Tracer

func init() {
	//conventional logging is 'always on'
	trLog = &TraceLog{
		logger: logging.Logger(),
	}
}

// InitTracing -
func InitTracing(serviceDisplay string) TraceLogging {
	if opentracingEnabled {
		cfg := &jaegerconfig.Configuration{
			Sampler: &jaegerconfig.SamplerConfig{
				Type:  "const",
				Param: 1,
			},
			Reporter: &jaegerconfig.ReporterConfig{
				LogSpans: true,
			},
		}

		tracer, closer, err := cfg.New(serviceDisplay, jaegerconfig.Logger(jaeger.StdLogger))
		if err != nil {
			panic(fmt.Sprintf("ERROR: cannot init Jaeger: %v\n", err))
		}

		opentracing.SetGlobalTracer(tracer)

		return TraceLogging{
			Closer: closer,
		}
	}

	return NoopTraceLogging
}

// TraceToBase64 -
func TraceToBase64(trace Tracer) (string, error) {

	switch t := trace.(type) {
	case TraceSpan:
		span := t.otSpan
		ext.SpanKindRPCClient.Set(span)

		textCarrier := opentracing.TextMapCarrier{}
		span.Tracer().Inject(span.Context(), opentracing.TextMap, textCarrier)

		var b []byte
		var err error
		if b, err = json.Marshal(textCarrier); err != nil {
			t.LogStringField("Error", err.Error())
			return "", err
		}
		result := base64.StdEncoding.EncodeToString(b)
		return result, nil
	default:
		return "", fmt.Errorf("unknown trace type: %v", t)
	}
}

// Base64ToTrace -
func Base64ToTrace(base64str, startSpanMsg string) (Tracer, error) {
	jsonB, err := base64.StdEncoding.DecodeString(base64str)
	textCarrier := opentracing.TextMapCarrier{}
	err = json.Unmarshal(jsonB, &textCarrier)
	if err != nil {
		return nil, err
	}

	tracer := opentracing.GlobalTracer()
	spanCtx, _ := tracer.Extract(opentracing.TextMap, textCarrier)
	span := tracer.StartSpan(startSpanMsg, ext.RPCServerOption(spanCtx))

	return TraceSpan{
			otSpan:   span,
			traceLog: trLog,
		},
		nil
}

// ContextWithSpanToBase64 - base64 is encoding applied to SpanContext
func ContextWithSpanToBase64(ctx context.Context) (string, error) {
	span := opentracing.SpanFromContext(ctx)
	if span != nil {
		ext.SpanKindRPCClient.Set(span)

		textCarrier := opentracing.TextMapCarrier{}
		span.Tracer().Inject(span.Context(), opentracing.TextMap, textCarrier)

		var b []byte
		var err error
		if b, err = json.Marshal(textCarrier); err != nil {
			trLog.LogErrorField("error marshalling span context to json", err)
			return "", err
		}
		result := base64.StdEncoding.EncodeToString(b)
		return result, nil
	} else {
		return "", fmt.Errorf("unable to serialize SpanContext, no span in context: %v", ctx)
	}
}

// ContextWithSpan - return context based on type of trace
func ContextWithSpan(ctx context.Context, trace Tracer) (context.Context, bool) {
	switch t := trace.(type) {
	case TraceSpan:
		return opentracing.ContextWithSpan(ctx, t.otSpan), true
	default:
		return ctx, false
	}
}

// StartTraceFromContext - return TraceSpan if opentracting is enabled else just TraceLog to log to file; if opentracing is enabled, we do both: opentracing and logging to file
func StartTraceFromContext(ctx context.Context, msg string) (Tracer, context.Context) {
	// is opentracing available/enabled? if so, return TraceSpan else a logger
	if opentracingEnabled {
		span, ctxWithSpan := opentracing.StartSpanFromContext(ctx, msg)

		return TraceSpan{
			otSpan:   span,
			traceLog: trLog,
		}, ctxWithSpan
	}
	return trLog, ctx
}

func spanFromMetadataOrNew(openTracingContext, msg string) (Tracer, context.Context) {
	var span Tracer
	var ctxWithSpan context.Context
	if len(openTracingContext) > 0 {
		span, _ = Base64ToTrace(openTracingContext, msg)
		ctxWithSpan, _ = ContextWithSpan(context.Background(), span)
	} else {
		span, ctxWithSpan = StartTraceFromContext(context.Background(), msg)
	}
	return span, ctxWithSpan
}

// IssueTrace -
func IssueTrace(openTracingContext, eventMsg, uuid, parentUUID string) context.Context {
	trace, ctxWithSpan := spanFromMetadataOrNew(openTracingContext, eventMsg)

	trace.LogStringField("event", eventMsg)
	trace.LogStringField("message.UUID", uuid)
	trace.LogStringField("message.Parent_UUID", parentUUID)
	trace.Finish()

	return ctxWithSpan
}

// IssueErrorTrace -
func IssueErrorTrace(opentracingContext string, err error, msg, uuid, parentUUID string) context.Context {
	trace, ctxWithSpan := spanFromMetadataOrNew(opentracingContext, "error")

	trace.LogErrorField("error", err)
	trace.LogStringField("error-info", msg)
	trace.LogStringField("message.UUID", uuid)
	trace.LogStringField("message.Parent_UUID", parentUUID)
	trace.Finish()

	return ctxWithSpan
}

// TraceLogging -
type TraceLogging struct {
	Closer io.Closer
}

// Close -
func (t TraceLogging) Close() {
	t.Closer.Close()
}

// NoopTraceLogging so the value under Closer key is never nil; this is the initial value of sidecar.traceLogging
// it allows call to Close in cases when sidecar is shutdown before registration completes and traceLogging is initialized
var NoopTraceLogging = TraceLogging{
	Closer: noopCloser{},
}

//Tracer - an interface to be implemented by TraceSpan and TraceLog types
type Tracer interface {
	LogStringField(key, value string)
	LogIntField(key string, value int)
	LogErrorField(key string, value error)
	Finish()
}

// TraceSpan - represents wrapper around opentracing; implements Tracer interface
type TraceSpan struct {
	otSpan   opentracing.Span
	traceLog Tracer
}

// TraceLog - represents wrapper around conventional logging; implements Tracer interface
// PLEASE NOTE: uses INFO level, since it corresponds the closest to the purpose of opentracing
type TraceLog struct {
	logger *zap.Logger
}

// Finish - TraceSpan implementation of Tracer interface
func (s TraceSpan) Finish() {
	s.otSpan.Finish()
}

// LogStringField -
func (s TraceSpan) LogStringField(key, value string) {
	s.otSpan.LogFields(
		opentracingLog.String(key, value),
	)
	s.traceLog.LogStringField(key, value)
}

// LogIntField - TraceSpan implementation of Tracer interface method
func (s TraceSpan) LogIntField(key string, value int) {
	s.otSpan.LogFields(
		opentracingLog.Int(key, value),
	)
	s.traceLog.LogIntField(key, value)
}

// LogErrorField - TraceSpan implementation of Tracer interface method
func (s TraceSpan) LogErrorField(key string, value error) {
	ext.Error.Set(s.otSpan, true)
	s.otSpan.LogFields(
		opentracingLog.Error(value),
	)
	s.traceLog.LogErrorField(key, value)
}

// LogStringField - TraceLog implementation of Tracer interface method
func (l TraceLog) LogStringField(key, value string) {
	l.logger.Info("tracing",
		zap.String(key, value),
	)
}

// LogIntField is TraceLog implementation of Tracer interface
func (l TraceLog) LogIntField(key string, value int) {
	l.logger.Info("tracing",
		zap.Int(key, value),
	)
}

// LogErrorField is TraceLog implementation of Tracer interface
func (l TraceLog) LogErrorField(key string, value error) {
	l.logger.Error("tracing",
		zap.Error(value),
	)
}

// Finish - in case of TraceLog implementation, there is nothing to do
func (l TraceLog) Finish() {
	//noop
}

type noopCloser struct {
}

func (n noopCloser) Close() error {
	return nil
}

// GetOpenTracingClientInterceptor -
func GetOpenTracingClientInterceptor() grpc.UnaryClientInterceptor {
	tracer := opentracing.GlobalTracer()
	return otgrpc.OpenTracingClientInterceptor(tracer)
}

// GetOpenTracingServerInterceptor -
func GetOpenTracingServerInterceptor() grpc.UnaryServerInterceptor {
	tracer := opentracing.GlobalTracer()
	return otgrpc.OpenTracingServerInterceptor(tracer)
}
