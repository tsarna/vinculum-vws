package server

import (
	"context"
	"testing"

	bus "github.com/tsarna/vinculum-bus"
	"go.opentelemetry.io/otel/metric/noop"
	"go.uber.org/zap"
)

// testEventBus implements a minimal EventBus for testing
type testEventBus struct{}

// EventBus methods
func (m *testEventBus) Start() error {
	return nil
}

func (m *testEventBus) Stop() error {
	return nil
}

func (m *testEventBus) Subscribe(ctx context.Context, topic string, subscriber bus.Subscriber) error {
	return nil
}

func (m *testEventBus) SubscribeFunc(ctx context.Context, topic string, receiver bus.EventReceiver) (bus.Subscriber, error) {
	return nil, nil
}

func (m *testEventBus) Unsubscribe(ctx context.Context, topic string, subscriber bus.Subscriber) error {
	return nil
}

func (m *testEventBus) UnsubscribeAll(ctx context.Context, subscriber bus.Subscriber) error {
	return nil
}

func (m *testEventBus) Publish(ctx context.Context, topic string, message any) error {
	return nil
}

func (m *testEventBus) PublishSync(ctx context.Context, topic string, message any) error {
	return nil
}

// Subscriber methods (EventBus extends Subscriber)
func (m *testEventBus) OnSubscribe(ctx context.Context, topic string) error {
	return nil
}

func (m *testEventBus) OnUnsubscribe(ctx context.Context, topic string) error {
	return nil
}

func (m *testEventBus) OnEvent(ctx context.Context, topic string, message any, fields map[string]string) error {
	return nil
}

func (m *testEventBus) PassThrough(msg bus.EventBusMessage) error {
	return nil
}

func TestListenerConfig_WithMeterProvider(t *testing.T) {
	logger := zap.NewNop()
	eventBus := &testEventBus{}
	mp := noop.NewMeterProvider()

	config := NewListener().
		WithEventBus(eventBus).
		WithLogger(logger).
		WithMeterProvider(mp)

	if config.meterProvider != mp {
		t.Error("Expected meter provider to be set")
	}

	// Test building listener with metrics
	listener, err := config.Build()
	if err != nil {
		t.Fatalf("Failed to build listener: %v", err)
	}

	if listener.metrics == nil {
		t.Error("Expected listener to have metrics initialized")
	}

	// Verify metrics are properly initialized
	if listener.metrics.activeConnections == nil {
		t.Error("Expected activeConnections metric to be initialized")
	}
}

func TestListenerConfig_WithoutMeterProvider(t *testing.T) {
	logger := zap.NewNop()
	eventBus := &testEventBus{}

	config := NewListener().
		WithEventBus(eventBus).
		WithLogger(logger)
		// No meter provider set

	listener, err := config.Build()
	if err != nil {
		t.Fatalf("Failed to build listener: %v", err)
	}

	// Metrics should be nil when no provider is set
	if listener.metrics != nil {
		t.Error("Expected listener metrics to be nil when no provider is set")
	}
}

func TestListenerConfig_MeterProviderInExample(t *testing.T) {
	// Test that the example in the documentation would work
	logger := zap.NewNop()
	eventBus := &testEventBus{}
	mp := noop.NewMeterProvider()

	listener, err := NewListener().
		WithEventBus(eventBus).
		WithLogger(logger).
		WithMeterProvider(mp).
		WithQueueSize(512).
		Build()

	if err != nil {
		t.Fatalf("Failed to build listener from example config: %v", err)
	}

	if listener == nil {
		t.Fatal("Expected listener to be created")
	}

	if listener.metrics == nil {
		t.Error("Expected listener to have metrics when provider is set")
	}
}
