package tracing

import (
	"context"
	"fmt"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	oteltrace "go.opentelemetry.io/otel/trace"
)

const (
	// Span names
	SpanNameRPCRequest        = "rpc.request"
	SpanNameSendRawTx         = "eth.sendRawTransaction"
	SpanNameSubmitTransaction = "eth.submitTransaction"
	SpanNameTxPoolAdd         = "txpool.add"
	SpanNameEngineAPI         = "engine.api"
	SpanNameGetPayload        = "engine.getPayload"
	SpanNameNewPayload        = "engine.newPayload"
	SpanNameForkchoiceUpdate  = "engine.forkchoiceUpdate"
	SpanNameTxForward         = "tx.forward"
	SpanNameBlockBuilding     = "block.building"
)

// Attribute keys following OpenTelemetry semantic conventions
const (
	// RPC semantic conventions (https://opentelemetry.io/docs/specs/semconv/rpc/)
	AttrRPCMethod    = "rpc.method"    // Standard: rpc.method
	AttrRPCSystem    = "rpc.system"   // Standard: rpc.system (e.g., "jsonrpc")
	AttrRPCService   = "rpc.service"  // Standard: rpc.service
	AttrRPCStatusCode = "rpc.status_code" // Standard: rpc.status_code
	
	// HTTP semantic conventions (https://opentelemetry.io/docs/specs/semconv/http/)
	AttrHTTPMethod      = "http.method"       // Standard: http.method
	AttrHTTPURL         = "http.url"          // Standard: http.url
	AttrHTTPStatusCode  = "http.status_code"  // Standard: http.status_code
	AttrHTTPUserAgent   = "http.user_agent"   // Standard: http.user_agent
	
	// Network semantic conventions (https://opentelemetry.io/docs/specs/semconv/network/)
	AttrNetPeerIP   = "net.peer.ip"   // Standard: net.peer.ip
	AttrNetPeerName = "net.peer.name" // Standard: net.peer.name
	AttrNetHostName = "net.host.name" // Standard: net.host.name
	
	// Error semantic conventions (https://opentelemetry.io/docs/specs/semconv/exceptions/)
	AttrErrorType    = "error.type"    // Standard: error.type
	AttrErrorMessage = "error.message" // Standard: error.message
	
	// Application-specific attributes (domain-specific, following dot notation)
	AttrTxHash      = "tx.hash"
	AttrTxFrom      = "tx.from"
	AttrTxTo        = "tx.to"
	AttrTxGas       = "tx.gas"
	AttrTxGasPrice  = "tx.gasPrice"
	AttrTxValue     = "tx.value"
	AttrBlockNumber = "block.number"
	AttrBlockHash   = "block.hash"
	AttrPayloadID   = "payload.id"
	AttrEngineMethod = "engine.method"
	AttrBackendURL  = "backend.url"
	AttrRequestID   = "request.id"
	
	// Legacy attribute names (kept for backward compatibility)
	AttrMethod      = AttrRPCMethod
	AttrUserAgent   = AttrHTTPUserAgent
	AttrRemoteAddr  = AttrNetPeerIP
	AttrErrorCode   = "rpc.error_code" // For RPC-specific error codes
)

// StartRPCSpan starts a span for an RPC request following OpenTelemetry semantic conventions
func StartRPCSpan(ctx context.Context, method string) (context.Context, oteltrace.Span) {
	attrs := []attribute.KeyValue{
		attribute.String(AttrRPCMethod, method),
		attribute.String(AttrRPCSystem, "jsonrpc"), // JSON-RPC is the RPC system
	}
	
	// Add service name if available
	if cfg := GetConfig(); cfg != nil && cfg.ServiceName != "" {
		attrs = append(attrs, attribute.String(AttrRPCService, cfg.ServiceName))
	}
	
	ctx, span := StartSpan(ctx, SpanNameRPCRequest,
		oteltrace.WithSpanKind(oteltrace.SpanKindServer),
		oteltrace.WithAttributes(attrs...),
	)
	return ctx, span
}

// StartSendRawTransactionSpan starts a span for sendRawTransaction
// Returns a no-op span if transaction tracing is not enabled
func StartSendRawTransactionSpan(ctx context.Context, txHash string) (context.Context, oteltrace.Span) {
	if !GetConfig().IsTransactionTracingEnabled() {
		return ctx, oteltrace.SpanFromContext(ctx)
	}
	ctx, span := StartSpan(ctx, SpanNameSendRawTx,
		oteltrace.WithSpanKind(oteltrace.SpanKindServer),
		oteltrace.WithAttributes(
			attribute.String(AttrTxHash, txHash),
		),
	)
	return ctx, span
}

// StartTxForwardSpan starts a span for transaction forwarding
// Returns a no-op span if transaction tracing is not enabled
func StartTxForwardSpan(ctx context.Context, txHash string, backendURL string) (context.Context, oteltrace.Span) {
	if !GetConfig().IsTransactionTracingEnabled() {
		return ctx, oteltrace.SpanFromContext(ctx)
	}
	ctx, span := StartSpan(ctx, SpanNameTxForward,
		oteltrace.WithSpanKind(oteltrace.SpanKindClient),
		oteltrace.WithAttributes(
			attribute.String(AttrTxHash, txHash),
			attribute.String(AttrBackendURL, backendURL),
		),
	)
	return ctx, span
}

// StartEngineAPISpan starts a span for Engine API calls
func StartEngineAPISpan(ctx context.Context, method string) (context.Context, oteltrace.Span) {
	ctx, span := StartSpan(ctx, SpanNameEngineAPI,
		oteltrace.WithSpanKind(oteltrace.SpanKindServer),
		oteltrace.WithAttributes(
			attribute.String(AttrEngineMethod, method),
		),
	)
	return ctx, span
}

// StartTxPoolSpan starts a span for transaction pool operations
// Returns a no-op span if transaction tracing is not enabled
func StartTxPoolSpan(ctx context.Context, txHash string) (context.Context, oteltrace.Span) {
	if !GetConfig().IsTransactionTracingEnabled() {
		return ctx, oteltrace.SpanFromContext(ctx)
	}
	ctx, span := StartSpan(ctx, SpanNameTxPoolAdd,
		oteltrace.WithAttributes(
			attribute.String(AttrTxHash, txHash),
		),
	)
	return ctx, span
}

// AddTransactionAttributes adds transaction-related attributes to a span
func AddTransactionAttributes(span oteltrace.Span, txHash, from, to string, gas uint64, gasPrice, value string) {
	if !IsEnabled() {
		return
	}
	
	attrs := []attribute.KeyValue{
		attribute.String(AttrTxHash, txHash),
	}
	
	if from != "" {
		attrs = append(attrs, attribute.String(AttrTxFrom, from))
	}
	if to != "" {
		attrs = append(attrs, attribute.String(AttrTxTo, to))
	}
	if gas > 0 {
		attrs = append(attrs, attribute.Int64(AttrTxGas, int64(gas)))
	}
	if gasPrice != "" {
		attrs = append(attrs, attribute.String(AttrTxGasPrice, gasPrice))
	}
	if value != "" {
		attrs = append(attrs, attribute.String(AttrTxValue, value))
	}
	
	span.SetAttributes(attrs...)
}

// AddBlockAttributes adds block-related attributes to a span
func AddBlockAttributes(span oteltrace.Span, blockNumber uint64, blockHash string) {
	if !IsEnabled() {
		return
	}
	
	attrs := []attribute.KeyValue{}
	if blockNumber > 0 {
		attrs = append(attrs, attribute.Int64(AttrBlockNumber, int64(blockNumber)))
	}
	if blockHash != "" {
		attrs = append(attrs, attribute.String(AttrBlockHash, blockHash))
	}
	
	span.SetAttributes(attrs...)
}

// AddHTTPAttributes adds HTTP-related attributes to a span following OpenTelemetry semantic conventions
func AddHTTPAttributes(span oteltrace.Span, req *http.Request) {
	if !IsEnabled() || req == nil {
		return
	}
	
	attrs := []attribute.KeyValue{}
	
	// HTTP method (standard semantic convention)
	if req.Method != "" {
		attrs = append(attrs, attribute.String(AttrHTTPMethod, req.Method))
	}
	
	// HTTP URL (standard semantic convention)
	if req.URL != nil {
		url := req.URL.String()
		if req.URL.Scheme == "" {
			// Reconstruct full URL if scheme is missing
			scheme := "http"
			if req.TLS != nil {
				scheme = "https"
			}
			url = scheme + "://" + req.Host + req.URL.Path
			if req.URL.RawQuery != "" {
				url += "?" + req.URL.RawQuery
			}
		}
		attrs = append(attrs, attribute.String(AttrHTTPURL, url))
	}
	
	// HTTP user agent (standard semantic convention)
	if req.UserAgent() != "" {
		attrs = append(attrs, attribute.String(AttrHTTPUserAgent, req.UserAgent()))
	}
	
	// Network peer information (standard semantic convention)
	if req.RemoteAddr != "" {
		// Try to extract IP from RemoteAddr (format: "ip:port" or "[ip]:port")
		attrs = append(attrs, attribute.String(AttrNetPeerIP, req.RemoteAddr))
	}
	
	// Host name (standard semantic convention)
	if req.Host != "" {
		attrs = append(attrs, attribute.String(AttrNetHostName, req.Host))
	}
	
	span.SetAttributes(attrs...)
}

// SetSpanError sets error information on a span following OpenTelemetry semantic conventions
func SetSpanError(span oteltrace.Span, err error) {
	if !IsEnabled() || err == nil {
		return
	}
	
	// Set error status (standard OpenTelemetry practice)
	span.SetStatus(codes.Error, err.Error())
	
	// Add error attributes following semantic conventions
	attrs := []attribute.KeyValue{
		attribute.String(AttrErrorMessage, err.Error()),
	}
	
	// Add error type if available
	if errType := fmt.Sprintf("%T", err); errType != "" {
		attrs = append(attrs, attribute.String(AttrErrorType, errType))
	}
	
	span.SetAttributes(attrs...)
}

// SetSpanErrorWithCode sets error information with a code on a span following OpenTelemetry semantic conventions
func SetSpanErrorWithCode(span oteltrace.Span, code int, message string) {
	if !IsEnabled() {
		return
	}
	
	// Set error status (standard OpenTelemetry practice)
	span.SetStatus(codes.Error, message)
	
	// Add error attributes following semantic conventions
	attrs := []attribute.KeyValue{
		attribute.String(AttrErrorMessage, message),
		attribute.Int(AttrErrorCode, code), // RPC-specific error code
	}
	
	span.SetAttributes(attrs...)
}

// InjectTraceContext injects trace context into HTTP headers for outgoing requests.
// This propagates the current trace context so downstream services can link traces.
func InjectTraceContext(ctx context.Context, req *http.Request) {
	if !IsEnabled() {
		return
	}
	
	// Get the global text map propagator (defaults to W3C TraceContext)
	propagator := otel.GetTextMapPropagator()
	if propagator == nil {
		// Fallback to W3C TraceContext if no propagator is set
		propagator = propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		)
	}
	
	// Inject trace context into HTTP headers
	propagator.Inject(ctx, propagation.HeaderCarrier(req.Header))
}

// ExtractTraceContextFromHeaders extracts trace context from HTTP headers for incoming requests.
// This allows geth to link its traces with traces from upstream services (like proxies).
func ExtractTraceContextFromHeaders(ctx context.Context, headers http.Header) context.Context {
	if !IsEnabled() {
		return ctx
	}
	
	// Get the global text map propagator (defaults to W3C TraceContext)
	propagator := otel.GetTextMapPropagator()
	if propagator == nil {
		// Fallback to W3C TraceContext if no propagator is set
		propagator = propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		)
	}
	
	// Extract trace context from HTTP headers
	// This will extract traceparent, tracestate, and baggage headers
	return propagator.Extract(ctx, propagation.HeaderCarrier(headers))
}

// RecordEvent records an event on a span
func RecordEvent(span oteltrace.Span, name string, attrs ...attribute.KeyValue) {
	if !IsEnabled() {
		return
	}
	
	span.AddEvent(name, oteltrace.WithAttributes(attrs...))
}

// SetSpanAttributes sets multiple attributes on a span
func SetSpanAttributes(span oteltrace.Span, attrs ...attribute.KeyValue) {
	if !IsEnabled() {
		return
	}
	
	span.SetAttributes(attrs...)
}

// SetHTTPSpanStatus sets HTTP status code on a span following OpenTelemetry semantic conventions
func SetHTTPSpanStatus(span oteltrace.Span, statusCode int) {
	if !IsEnabled() {
		return
	}
	
	attrs := []attribute.KeyValue{
		attribute.Int(AttrHTTPStatusCode, statusCode),
	}
	
	// Set span status based on HTTP status code (standard OpenTelemetry practice)
	if statusCode >= 400 {
		span.SetStatus(codes.Error, fmt.Sprintf("HTTP %d", statusCode))
	} else {
		span.SetStatus(codes.Ok, "")
	}
	
	span.SetAttributes(attrs...)
}

// SetRPCSpanStatus sets RPC status code on a span following OpenTelemetry semantic conventions
func SetRPCSpanStatus(span oteltrace.Span, statusCode string) {
	if !IsEnabled() {
		return
	}
	
	span.SetAttributes(
		attribute.String(AttrRPCStatusCode, statusCode),
	)
}