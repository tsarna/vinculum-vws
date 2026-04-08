package server_test

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	bus "github.com/tsarna/vinculum-bus"
	"github.com/tsarna/vinculum-bus/o11y"
	"github.com/tsarna/vinculum-vws/server"
	"go.opentelemetry.io/otel/metric/noop"
	"go.uber.org/zap"
)

// ExampleWebSocketMetrics demonstrates how to set up a WebSocket server with metrics collection.
func ExampleWebSocketMetrics() {
	// Create a logger
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	// Create an EventBus (using the default implementation)
	eventBus, err := bus.NewEventBus().
		WithLogger(logger).
		WithBufferSize(1000).
		Build()
	if err != nil {
		log.Fatalf("Build() returned error: %v", err)
	}
	defer eventBus.Stop()

	// Start the EventBus
	if err := eventBus.Start(); err != nil {
		log.Fatal(err)
	}

	// Create a standalone meter provider that publishes metrics to the bus
	mp, _ := o11y.NewStandaloneMeterProvider(eventBus, &o11y.StandaloneMetricsConfig{
		Interval:     30 * time.Second,
		MetricsTopic: "$metrics",
		ServiceName:  "websocket-server",
	})
	defer mp.Shutdown(context.Background()) //nolint:errcheck

	// Create the WebSocket listener with metrics
	listener, err := server.NewListener().
		WithEventBus(eventBus).
		WithLogger(logger).
		WithMeterProvider(mp).
		WithQueueSize(512).
		WithPingInterval(30 * time.Second).
		WithEventAuth(server.AllowTopicPrefix("client/")).
		Build()
	if err != nil {
		log.Fatal(err)
	}

	// Set up HTTP server
	http.Handle("/ws", listener)

	// Start server
	srv := &http.Server{
		Addr:    ":8080",
		Handler: http.DefaultServeMux,
	}

	fmt.Println("WebSocket server with metrics running on :8080")
	fmt.Println("Metrics will be published to topic '$metrics' every 30 seconds")
	fmt.Println("Connect to ws://localhost:8080/ws to test")

	// In a real application, you would handle graceful shutdown
	log.Fatal(srv.ListenAndServe())
}

// ExampleListener_metricsWithNoopProvider shows how to use a noop meter provider for testing.
func ExampleListener_metricsWithNoopProvider() {

	// Create logger and EventBus
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	eventBus, err := bus.NewEventBus().
		WithLogger(logger).
		WithBufferSize(1000).
		Build()
	if err != nil {
		log.Fatalf("Build() returned error: %v", err)
	}
	defer eventBus.Stop()

	if err := eventBus.Start(); err != nil {
		log.Fatal(err)
	}

	// Use noop meter provider (useful for testing)
	mp := noop.NewMeterProvider()

	// Create WebSocket listener
	listener, err := server.NewListener().
		WithEventBus(eventBus).
		WithLogger(logger).
		WithMeterProvider(mp).
		Build()
	if err != nil {
		log.Fatal(err)
	}

	// Set up HTTP server
	http.Handle("/ws", listener)

	fmt.Println("WebSocket server with noop metrics provider running on :8080")

	// Start server
	srv := &http.Server{
		Addr:    ":8080",
		Handler: http.DefaultServeMux,
	}

	log.Fatal(srv.ListenAndServe())
}

// ExampleListener_metricsDisabled shows how to run the WebSocket server without metrics.
func ExampleListener_metricsDisabled() {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	eventBus, err := bus.NewEventBus().
		WithLogger(logger).
		WithBufferSize(1000).
		Build()
	if err != nil {
		log.Fatalf("Build() returned error: %v", err)
	}
	defer eventBus.Stop()

	if err := eventBus.Start(); err != nil {
		log.Fatal(err)
	}

	// Create WebSocket listener without meter provider
	listener, err := server.NewListener().
		WithEventBus(eventBus).
		WithLogger(logger).
		// No WithMeterProvider() call - metrics will be disabled
		Build()
	if err != nil {
		log.Fatal(err)
	}

	http.Handle("/ws", listener)

	fmt.Println("WebSocket server running without metrics on :8080")

	srv := &http.Server{
		Addr:    ":8080",
		Handler: http.DefaultServeMux,
	}

	log.Fatal(srv.ListenAndServe())
}
