package vws

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

// InjectTrace injects trace context from ctx into msg.Headers using the global propagator.
// If no propagator is configured (the default no-op), this is a no-op and Headers is unchanged.
func InjectTrace(ctx context.Context, msg *WireMessage) {
	carrier := propagation.MapCarrier{}
	otel.GetTextMapPropagator().Inject(ctx, carrier)
	if len(carrier) > 0 {
		msg.Headers = map[string]string(carrier)
	}
}

// ExtractTrace extracts trace context from msg.Headers and returns a derived context.
// If Headers is empty or no propagator is configured, the original ctx is returned unchanged.
func ExtractTrace(ctx context.Context, msg WireMessage) context.Context {
	if len(msg.Headers) == 0 {
		return ctx
	}
	return otel.GetTextMapPropagator().Extract(ctx, propagation.MapCarrier(msg.Headers))
}

// HeadersFromContext extracts propagation headers from ctx as a map, or nil if empty.
// This is used by the server's pre-allocated eventMsg map to avoid creating a WireMessage.
func HeadersFromContext(ctx context.Context) map[string]string {
	carrier := propagation.MapCarrier{}
	otel.GetTextMapPropagator().Inject(ctx, carrier)
	if len(carrier) == 0 {
		return nil
	}
	return map[string]string(carrier)
}
