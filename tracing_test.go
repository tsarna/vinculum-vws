package vws_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"

	vws "github.com/tsarna/vinculum-vws"
)

const (
	testTraceIDHex = "4bf92f3577b34da6a3ce929d0e0e4736"
	testSpanIDHex  = "00f067aa0ba902b7"
)

func makeSpanContext(t *testing.T) trace.SpanContext {
	t.Helper()
	traceID, err := trace.TraceIDFromHex(testTraceIDHex)
	require.NoError(t, err)
	spanID, err := trace.SpanIDFromHex(testSpanIDHex)
	require.NoError(t, err)
	return trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    traceID,
		SpanID:     spanID,
		TraceFlags: trace.FlagsSampled,
		Remote:     true,
	})
}

func TestInjectTrace_NoOp(t *testing.T) {
	// Default no-op propagator: no headers should be set
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator())
	defer otel.SetTextMapPropagator(nil)

	msg := &vws.WireMessage{Topic: "test"}
	vws.InjectTrace(context.Background(), msg)

	assert.Nil(t, msg.Headers)
}

func TestInjectTrace_WithTraceContext(t *testing.T) {
	otel.SetTextMapPropagator(propagation.TraceContext{})
	defer otel.SetTextMapPropagator(nil)

	sc := makeSpanContext(t)
	ctx := trace.ContextWithSpanContext(context.Background(), sc)

	msg := &vws.WireMessage{Topic: "test"}
	vws.InjectTrace(ctx, msg)

	require.NotNil(t, msg.Headers)
	assert.Contains(t, msg.Headers, "traceparent")
	assert.Contains(t, msg.Headers["traceparent"], testTraceIDHex)
}

func TestExtractTrace_NoHeaders(t *testing.T) {
	otel.SetTextMapPropagator(propagation.TraceContext{})
	defer otel.SetTextMapPropagator(nil)

	msg := vws.WireMessage{Topic: "test"}
	result := vws.ExtractTrace(context.Background(), msg)

	// No headers → no valid span context
	assert.False(t, trace.SpanFromContext(result).SpanContext().IsValid())
}

func TestExtractTrace_WithHeaders(t *testing.T) {
	otel.SetTextMapPropagator(propagation.TraceContext{})
	defer otel.SetTextMapPropagator(nil)

	msg := vws.WireMessage{
		Topic: "test",
		Headers: map[string]string{
			"traceparent": "00-" + testTraceIDHex + "-" + testSpanIDHex + "-01",
		},
	}

	ctx := vws.ExtractTrace(context.Background(), msg)
	sc := trace.SpanFromContext(ctx).SpanContext()

	assert.True(t, sc.IsValid())
	assert.True(t, sc.IsRemote())
	assert.Equal(t, testTraceIDHex, sc.TraceID().String())
	assert.Equal(t, testSpanIDHex, sc.SpanID().String())
}

func TestInjectExtractRoundTrip(t *testing.T) {
	otel.SetTextMapPropagator(propagation.TraceContext{})
	defer otel.SetTextMapPropagator(nil)

	original := makeSpanContext(t)
	sendCtx := trace.ContextWithSpanContext(context.Background(), original)

	msg := &vws.WireMessage{Topic: "test"}
	vws.InjectTrace(sendCtx, msg)

	receiveCtx := vws.ExtractTrace(context.Background(), *msg)
	extracted := trace.SpanFromContext(receiveCtx).SpanContext()

	assert.Equal(t, original.TraceID(), extracted.TraceID())
	assert.Equal(t, original.SpanID(), extracted.SpanID())
	assert.Equal(t, original.TraceFlags(), extracted.TraceFlags())
}

func TestHeadersFromContext_NoOp(t *testing.T) {
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator())
	defer otel.SetTextMapPropagator(nil)

	result := vws.HeadersFromContext(context.Background())
	assert.Nil(t, result)
}

func TestHeadersFromContext_WithTrace(t *testing.T) {
	otel.SetTextMapPropagator(propagation.TraceContext{})
	defer otel.SetTextMapPropagator(nil)

	sc := makeSpanContext(t)
	ctx := trace.ContextWithSpanContext(context.Background(), sc)

	headers := vws.HeadersFromContext(ctx)
	require.NotNil(t, headers)
	assert.Contains(t, headers, "traceparent")
	assert.Contains(t, headers["traceparent"], testTraceIDHex)
}
